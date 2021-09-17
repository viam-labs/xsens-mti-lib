package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/edaniels/golog"
	"go.uber.org/multierr"
	"go.viam.com/core/config"
	"go.viam.com/core/grpc/server"
	pb "go.viam.com/core/proto/api/v1"
	"go.viam.com/core/rlog"
	robotimpl "go.viam.com/core/robot/impl"

	"go.viam.com/mti/serial"
	"go.viam.com/utils"
	goutils "go.viam.com/utils"
	rpcserver "go.viam.com/utils/rpc/server"
	"goji.io"
	"goji.io/pat"
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
	Port       goutils.NetPortFlag `flag:"0"`
	DevicePath string              `flag:"device,default=/dev/ttyUSB0,usage=device path"`
	DeviceID   string              `flag:"device-id,default=02782090,usage=device id"`
}

func mainWithArgs(ctx context.Context, args []string, logger golog.Logger) error {
	var argsParsed Arguments
	if err := goutils.ParseFlags(args, &argsParsed); err != nil {
		return err
	}
	if argsParsed.Port == 0 {
		argsParsed.Port = goutils.NetPortFlag(defaultPort)
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
	r.AddSensor(sensor, config.Component{Name: "compass"})

	if err := rpcServer.RegisterServiceServer(
		ctx,
		&pb.RobotService_ServiceDesc,
		server.New(r),
		pb.RegisterRobotServiceHandlerFromEndpoint,
	); err != nil {
		return err
	}

	mux := goji.NewMux()
	mux.Handle(pat.New("/api/*"), http.StripPrefix("/api", rpcServer.GatewayHandler()))
	mux.Handle(pat.New("/*"), rpcServer.GRPCHandler())

	httpServer, err := goutils.NewPlainTextHTTP2Server(mux)
	if err != nil {
		return err
	}
	httpServer.Addr = listener.Addr().String()

	stopped := make(chan struct{})
	defer func() {
		<-stopped
	}()
	goutils.PanicCapturingGo(func() {
		defer func() {
			close(stopped)
		}()
		<-ctx.Done()
		defer func() {
			if err := httpServer.Shutdown(context.Background()); err != nil {
				logger.Errorw("error shutting down", "error", err)
			}
		}()
		if err := rpcServer.Stop(); err != nil {
			logger.Errorw("error stopping rpc server", "error", err)
		}
	})
	goutils.PanicCapturingGo(func() {
		if err := rpcServer.Start(); err != nil {
			logger.Errorw("error starting rpc server", "error", err)
		}
	})

	logger.Debugw("serving", "url", fmt.Sprintf("http://%s", listener.Addr().String()))
	if err := httpServer.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}
