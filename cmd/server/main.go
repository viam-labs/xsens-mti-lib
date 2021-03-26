package main

import (
	"context"
	"fmt"
	"net"

	"go.viam.com/mti/serial"

	"github.com/edaniels/golog"
	"go.uber.org/multierr"
	pb "go.viam.com/robotcore/proto/sensor/compass/v1"
	"go.viam.com/robotcore/rlog"
	"go.viam.com/robotcore/rpc"
	"go.viam.com/robotcore/sensor/compass"
	"go.viam.com/robotcore/utils"
)

func main() {
	utils.ContextualMain(mainWithArgs, logger)
}

var (
	defaultPort = 4444
	logger      = rlog.Logger.Named("server")
)

// Arguments for the command.
type Arguments struct {
	Port       utils.NetPortFlag `flag:"0"`
	DevicePath string            `flag:"device,default=/dev/ttyUSB0,usage=device path"`
	DeviceID   string            `flag:"device-id,default=02782090,usage=device id"`
}

func mainWithArgs(ctx context.Context, args []string, logger golog.Logger) error {
	var argsParsed Arguments
	if err := utils.ParseFlags(args, &argsParsed); err != nil {
		return err
	}
	if argsParsed.Port == 0 {
		argsParsed.Port = utils.NetPortFlag(defaultPort)
	}

	return runServer(ctx, int(argsParsed.Port), argsParsed.DevicePath, argsParsed.DeviceID, logger)
}

func runServer(ctx context.Context, port int, devicePath, deviceID string, logger golog.Logger) (err error) {
	sensor, err := serial.NewDevice(deviceID, devicePath, 115200)
	if err != nil {
		return err
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return err
	}

	rpcServer, err := rpc.NewServer()
	if err != nil {
		return err
	}
	defer func() {
		err = multierr.Combine(err, rpcServer.Stop())
	}()

	if err := rpcServer.RegisterServiceServer(
		ctx,
		&pb.CompassService_ServiceDesc,
		compass.NewServer(sensor),
		pb.RegisterCompassServiceHandlerFromEndpoint,
	); err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		if err := rpcServer.Stop(); err != nil {
			panic(err)
		}
	}()
	utils.ContextMainReadyFunc(ctx)()
	return rpcServer.Serve(listener)
}
