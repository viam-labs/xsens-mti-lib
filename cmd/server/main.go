package main

import (
	"context"
	"fmt"
	"net"

	"github.com/edaniels/golog"
	"go.uber.org/multierr"
	"go.viam.com/core/config"
	"go.viam.com/core/grpc/server"
	pb "go.viam.com/core/proto/api/v1"
	"go.viam.com/core/rlog"
	robotimpl "go.viam.com/core/robot/impl"
	"go.viam.com/mti/serial"
	"go.viam.com/utils"
	rpcserver "go.viam.com/utils/rpc/server"
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
	sensor, err := serial.NewCompass(deviceID, devicePath, 115200)
	if err != nil {
		return err
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return err
	}

	rpcServer, err := rpcserver.New(logger)
	if err != nil {
		return err
	}
	defer func() {
		err = multierr.Combine(err, rpcServer.Stop())
	}()

	r, err := robotimpl.New(ctx, &config.Config{}, logger)
	if err != nil {
		return err
	}
	r.AddSensor(sensor, config.Component{})

	if err := rpcServer.RegisterServiceServer(
		ctx,
		&pb.RobotService_ServiceDesc,
		server.New(r),
		pb.RegisterRobotServiceHandlerFromEndpoint,
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
