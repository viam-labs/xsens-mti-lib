package xsense

import (
	"context"
	"io"
	"sync"

	"github.com/edaniels/golog"
	"github.com/golang/geo/r3"
	geo "github.com/kellydunn/golang-geo"
	"github.com/pkg/errors"
	"go.viam.com/utils"

	mtilib "github.com/viam-labs/xsens-mti-lib/serial"
	"go.viam.com/rdk/components/movementsensor"
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/resource"
	"go.viam.com/rdk/spatialmath"
	rutils "go.viam.com/rdk/utils"
)

var Model = resource.NewModel("viam", "sensor", "mti-xsense-200")
var baudRateList = []uint{115200}

func init() {
	resource.RegisterComponent(
		movementsensor.API,
		Model,
		resource.Registration[sensor.Sensor, *Config]{
			Constructor: func(
				ctx context.Context,
				deps resource.Dependencies,
				conf resource.Config,
				logger golog.Logger,
			) (sensor.Sensor, error) {
				newConf, err := resource.NativeConfig[*Config](conf)
				if err != nil {
					return nil, err
				}
				return newXsens(ctx, deps, conf.ResourceName(), newConf, logger)
			},
		})
}

type Config struct {
	SerialPath     string `json:"serial_path"`
	SerialBaudRate int    `json:"serial_baud_rate,omitempty"`
	DeviceID       string `json:"device_id"`
}

// Validate ensures all parts of the config are valid.
func (cfg *Config) Validate(path string) ([]string, error) {
	var deps []string
	if cfg.SerialPath == "" {
		return nil, utils.NewConfigValidationFieldRequiredError(path, "serial_path")
	}
	// Validating baud rate
	if !rutils.ValidateBaudRate(baudRateList, int(cfg.SerialBaudRate)) {
		return nil, utils.NewConfigValidationError(path, errors.Errorf("Baud rate is not in %v", baudRateList))
	}

	return deps, nil
}

type xsense struct {
	resource.Named
	resource.AlwaysRebuild
	magnetometer            r3.Vector
	compassheading          float64
	numBadReadings          uint32
	err                     movementsensor.LastError
	mu                      sync.Mutex
	port                    io.ReadWriteCloser
	cancelFunc              func()
	cancelCtx               context.Context
	activeBackgroundWorkers sync.WaitGroup
	logger                  golog.Logger
	imu                     movementsensor.MovementSensor
}

// Close
func (i *xsense) Close(ctx context.Context) error {
	return i.imu.Close(ctx)
}

// CompassHeading
func (i *xsense) CompassHeading(ctx context.Context, extra map[string]interface{}) (float64, error) {
	return i.imu.CompassHeading(ctx, extra)
}

// Accuracy unimplemented
func (i *xsense) Accuracy(ctx context.Context, extra map[string]interface{}) (map[string]float32, error) {
	i.mu.Lock()
	defer i.mu.Unlock()
	return nil, nil
}

// AngularVelocity unimplemented
func (i *xsense) AngularVelocity(ctx context.Context, extra map[string]interface{}) (spatialmath.AngularVelocity, error) {
	i.mu.Lock()
	defer i.mu.Unlock()
	return spatialmath.AngularVelocity{}, nil
}

// LinearAcceleration unimplemented
func (i *xsense) LinearAcceleration(ctx context.Context, extra map[string]interface{}) (r3.Vector, error) {
	i.mu.Lock()
	defer i.mu.Unlock()
	return r3.Vector{}, nil
}

// LinearVelocity unimplemented
func (i *xsense) LinearVelocity(ctx context.Context, extra map[string]interface{}) (r3.Vector, error) {
	i.mu.Lock()
	defer i.mu.Unlock()
	return r3.Vector{}, nil
}

// Orientation unimplemented
func (i *xsense) Orientation(ctx context.Context, extra map[string]interface{}) (spatialmath.Orientation, error) {
	i.mu.Lock()
	defer i.mu.Unlock()
	return spatialmath.NewZeroOrientation(), nil
}

// Position unimplemented
func (i *xsense) Position(ctx context.Context, extra map[string]interface{}) (*geo.Point, float64, error) {
	i.mu.Lock()
	defer i.mu.Unlock()
	return nil, 0, nil
}

// Properties
func (mti *xsense) Properties(ctx context.Context, extra map[string]interface{}) (*movementsensor.Properties, error) {
	mti.mu.Lock()
	defer mti.mu.Unlock()
	return &movementsensor.Properties{
		CompassHeadingSupported: true,
	}, nil
}

// Readings
func (i *xsense) Readings(ctx context.Context, extra map[string]interface{}) (map[string]interface{}, error) {
	readings := make(map[string]interface{})
	return readings, nil
}

func newXsens(
	ctx context.Context,
	deps resource.Dependencies,
	name resource.Name,
	newConf *Config,
	logger golog.Logger,
) (movementsensor.MovementSensor, error) {
	return mtilib.NewCompass(newConf.DeviceID, newConf.SerialPath, newConf.SerialBaudRate)
}
