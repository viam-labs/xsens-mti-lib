package main

import (
	"context"
	"xsens"

	"github.com/edaniels/golog"
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/module"
	"go.viam.com/utils"
)

func main() {
	utils.ContextualMain(mainWithArgs, golog.NewDevelopmentLogger("mti-xsense"))
}

func mainWithArgs(ctx context.Context, args []string, logger golog.Logger) error {
	imu, err := module.NewModuleFromArgs(ctx, logger)
	if err != nil {
		return err
	}

	imu.AddModelFromRegistry(ctx, sensor.API, xsens.Model)

	err = imu.Start(ctx)
	defer imu.Close(ctx)
	if err != nil {
		return err
	}

	<-ctx.Done()
	return nil
}
