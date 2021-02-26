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

	mtiserial "github.com/viamrobotics/mti/serial"
	mtiws "github.com/viamrobotics/mti/ws"
	"go.viam.com/robotcore/sensor/compass"

	"github.com/edaniels/golog"
	"github.com/edaniels/wsapi"
	"nhooyr.io/websocket"
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

	sensor, err := mtiserial.NewDevice("02782090", "/dev/ttyUSB0", 115200)
	if err != nil {
		golog.Global.Fatal(err)
	}

	httpServer := &http.Server{
		Addr:           fmt.Sprintf(":%d", port),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	httpServer.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			golog.Global.Error("error making websocket connection", "error", err)
			return
		}
		defer conn.Close(websocket.StatusNormalClosure, "")

		for {
			select {
			case <-r.Context().Done():
				return
			default:
			}

			cmd, err := wsapi.ReadCommand(r.Context(), conn)
			if err != nil {
				golog.Global.Errorw("error reading command", "error", err)
				return
			}
			result, err := processCommand(r.Context(), cmd, sensor)
			if err != nil {
				resp := wsapi.NewErrorResponse(err)
				if err := wsapi.WriteJSONResponse(r.Context(), resp, conn); err != nil {
					golog.Global.Errorw("error writing", "error", err)
					continue
				}
				continue
			}
			if err := wsapi.WriteJSONResponse(r.Context(), wsapi.NewSuccessfulResponse(result), conn); err != nil {
				golog.Global.Errorw("error writing", "error", err)
				continue
			}
		}
	})

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

func processCommand(ctx context.Context, cmd *wsapi.Command, sensor compass.Device) (interface{}, error) {
	switch cmd.Name {
	case mtiws.CommandHeading:
		return sensor.Heading(ctx)
	default:
		return nil, fmt.Errorf("unknown command %s", cmd.Name)
	}
}
