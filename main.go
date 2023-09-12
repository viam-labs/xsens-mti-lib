package main

import (
	"context"

	"github.com/edaniels/golog"
	"github.com/viam-labs/xsens-mti-lib/xsens"
	"go.viam.com/rdk/components/movementsensor"
	"go.viam.com/rdk/module"
	"go.viam.com/utils"
)

func main() {
	utils.ContextualMain(mainWithArgs, golog.NewDevelopmentLogger("mti-xsens"))
}

func mainWithArgs(ctx context.Context, args []string, logger golog.Logger) error {
	imu, err := module.NewModuleFromArgs(ctx, logger)
	if err != nil {
		return err
	}

	imu.AddModelFromRegistry(ctx, movementsensor.API, xsens.Model)

	err = imu.Start(ctx)
	defer imu.Close(ctx)
	if err != nil {
		return err
	}

	<-ctx.Done()
	return nil
}
