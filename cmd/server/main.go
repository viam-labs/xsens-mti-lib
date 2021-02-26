package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"go.viam.com/mti/serial"

	"github.com/edaniels/golog"
	"github.com/edaniels/wsapi"
	"go.viam.com/robotcore/sensor/compass"
)

func main() {
	flag.Parse()

	port := 4444
	if flag.NArg() >= 1 {
		portParsed, err := strconv.ParseInt(flag.Arg(0), 10, 32)
		if err != nil {
			golog.Global.Fatal(err)
		}
		port = int(portParsed)
	}

	sensor, err := serial.NewDevice("02782090", "/dev/ttyUSB0", 115200)
	if err != nil {
		golog.Global.Fatal(err)
	}

	httpServer := &http.Server{
		Addr:           fmt.Sprintf(":%d", port),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	wsServer := wsapi.NewServer()
	registerCommands(wsServer, sensor)
	httpServer.Handler = wsServer.HTTPHandler()

	errChan := make(chan error, 1)
	go func() {
		golog.Global.Infow("listening", "url", fmt.Sprintf("http://localhost:%d", port), "port", port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	select {
	case err := <-errChan:
		golog.Global.Errorw("failed to serve", "error", err)
	case <-sig:
	}

	if err := httpServer.Shutdown(context.Background()); err != nil {
		golog.Global.Fatal(err)
	}
}

func registerCommands(server wsapi.Server, sensor compass.Device) {
	server.RegisterCommand(compass.WSCommandHeading, wsapi.CommandHandlerFunc(func(ctx context.Context, cmd *wsapi.Command) (interface{}, error) {
		return sensor.Heading(ctx)
	}))
}
