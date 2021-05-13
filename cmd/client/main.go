package main

import (
	"context"
	"errors"
	"io"
	"os"
	"os/signal"
	"time"

	"go.viam.com/core/rlog"
	"go.viam.com/core/sensor/compass/client"
	"go.viam.com/core/utils"

	"github.com/edaniels/golog"
	"go.uber.org/multierr"
)

func main() {
	utils.ContextualMain(mainWithArgs, logger)
}

var logger = rlog.Logger.Named("client")

// Arguments for the command.
type Arguments struct {
	DeviceAddress string `flag:"device,required,default=localhost:4444,usage=device address"`
}

func mainWithArgs(ctx context.Context, args []string, logger golog.Logger) error {
	var argsParsed Arguments
	if err := utils.ParseFlags(args, &argsParsed); err != nil {
		return err
	}

	return runClient(ctx, argsParsed.DeviceAddress, logger)
}

func runClient(ctx context.Context, deviceAddress string, logger golog.Logger) (err error) {
	client, err := client.New(ctx, deviceAddress, logger)
	if err != nil {
		return err
	}

	defer func() {
		err = multierr.Combine(err, utils.TryClose(client))
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	utils.ContextMainReadyFunc(ctx)()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}

		heading, err := client.Heading(ctx)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		logger.Infow("heading", "data", heading)
	}
}
