package serial

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/edaniels/golog"
	"github.com/golang/geo/r3"
	geo "github.com/kellydunn/golang-geo"
	"github.com/viam-labs/xsens-mti-lib/gen"
	"go.viam.com/rdk/components/movementsensor"
	"go.viam.com/rdk/resource"
	"go.viam.com/rdk/spatialmath"
)

type Compass struct {
	control   gen.XsControl
	device    gen.XSDevice
	callback  gen.CallbackHandler
	heading   atomic.Value
	closeCh   chan struct{}
	closeOnce sync.Once
	mu        sync.Mutex
}

func NewCompass(deviceID string, path string, baudRate int) (movementsensor.MovementSensor, error) {
	control := gen.XsControlConstruct()

	portInfoArray := gen.XSScannerScanPorts()
	portInfoArrayPtr := gen.SwigcptrXsArrayXsPortInfo(portInfoArray.Swigcptr())

	if portInfoArrayPtr.Size() == 0 {
		return nil, errors.New("no mti device found")
	}
	mtPort := portInfoArrayPtr.First()

	foundPath := mtPort.PortName().ToStdString()
	golog.Global().Infof("found device",
		"id", mtPort.DeviceId().ToString().ToStdString(),
		"port", foundPath,
		"baudrate", mtPort.Baudrate(),
	)
	if foundPath != path {
		return nil, fmt.Errorf("found device at %q but not %q", foundPath, path)
	}

	var useBaudRate gen.XsBaudRate
	switch baudRate {
	case 115200:
		useBaudRate = gen.XBR_115k2
	default:
		return nil, fmt.Errorf("unknown baudrate %d", baudRate)
	}

	pathStr := gen.NewXSString(path)
	defer gen.DeleteXSString(pathStr)
	if !control.OpenPort(pathStr, useBaudRate) {
		defer control.Destruct()
		return nil, errors.New("failed to open port")
	}

	devID := gen.NewXSDeviceId()
	defer gen.DeleteXSDeviceId(devID)
	devIDStr := gen.NewXSString(deviceID)
	defer gen.DeleteXSString(devIDStr)
	devID.FromString(devIDStr)

	device := control.Device(devID)
	if device.Swigcptr() == 0 {
		defer control.Destruct()
		return nil, errors.New("expected device")
	}

	device.SetDeviceOptionFlags(gen.XDOF_EnableContinuousZRU, gen.XDOF_None)

	callback := gen.NewCallbackHandler()
	gen.AddCallbackHandler(callback, device)

	if !device.GotoMeasurement() {
		return nil, errors.New("failed to go to measurement mode")
	}

	c := &Compass{
		control:  control,
		device:   device,
		callback: callback,
	}
	c.heading.Store(math.NaN())

	c.closeCh = make(chan struct{})
	go func() {
		ticker := time.NewTicker(time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-c.closeCh:
				return
			case <-ticker.C:
			}

			if callback.PacketAvailable() {
				packet := callback.GetNextPacket()
				if packet.ContainsOrientation() {
					euler := packet.OrientationEuler()
					if yaw := euler.Yaw(); !math.IsNaN(yaw) {
						c.heading.Store(yaw)
					}
				}
			}

		}
	}()
	return c, nil
}

func (c *Compass) CompassHeading(ctx context.Context, extra map[string]interface{}) (float64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.heading.Load().(float64), nil
}

func (c *Compass) Close(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closeOnce.Do(func() {
		close(c.closeCh)
		defer c.control.Destruct()
		defer gen.DeleteCallbackHandler(c.callback)
	})
	return nil
}

func (c *Compass) Accuracy(ctx context.Context, extra map[string]interface{}) (map[string]float32, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return nil, nil
}

// AngularVelocity unimplemented
func (c *Compass) AngularVelocity(ctx context.Context, extra map[string]interface{}) (spatialmath.AngularVelocity, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return spatialmath.AngularVelocity{}, nil
}

// LinearAcceleration unimplemented
func (c *Compass) LinearAcceleration(ctx context.Context, extra map[string]interface{}) (r3.Vector, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return r3.Vector{}, nil
}

// LinearVelocity unimplemented
func (c *Compass) LinearVelocity(ctx context.Context, extra map[string]interface{}) (r3.Vector, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return r3.Vector{}, nil
}

// Orientation unimplemented
func (c *Compass) Orientation(ctx context.Context, extra map[string]interface{}) (spatialmath.Orientation, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return spatialmath.NewZeroOrientation(), nil
}

// Position unimplemented
func (c *Compass) Position(ctx context.Context, extra map[string]interface{}) (*geo.Point, float64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return nil, 0, nil
}

// Properties
func (c *Compass) Properties(ctx context.Context, extra map[string]interface{}) (*movementsensor.Properties, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return &movementsensor.Properties{
		CompassHeadingSupported: true,
	}, nil
}

// Readings
func (c *Compass) Readings(ctx context.Context, extra map[string]interface{}) (map[string]interface{}, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	readings := make(map[string]interface{})
	return readings, nil
}

// DoCommand implements movementsensor.MovementSensor.
func (*Compass) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	return nil, nil
}

// Name implements movementsensor.MovementSensor.
func (*Compass) Name() resource.Name {
	panic("unimplemented")
}

// Reconfigure implements movementsensor.MovementSensor.
func (*Compass) Reconfigure(ctx context.Context, deps resource.Dependencies, conf resource.Config) error {
	panic("unimplemented")
}
