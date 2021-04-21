package serial

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"go.viam.com/mti/gen"

	"github.com/edaniels/golog"
	"go.viam.com/robotcore/sensor"
	"go.viam.com/robotcore/sensor/compass"
)

type Device struct {
	control   gen.XsControl
	device    gen.XSDevice
	callback  gen.CallbackHandler
	heading   atomic.Value
	closeCh   chan struct{}
	closeOnce sync.Once
}

func NewDevice(deviceID string, path string, baudRate int) (compass.Device, error) {
	control := gen.XsControlConstruct()

	portInfoArray := gen.XSScannerScanPorts()
	portInfoArrayPtr := gen.SwigcptrXsArrayXsPortInfo(portInfoArray.Swigcptr())

	if portInfoArrayPtr.Size() == 0 {
		return nil, errors.New("no mti device found")
	}
	mtPort := portInfoArrayPtr.First()

	foundPath := mtPort.PortName().ToStdString()
	golog.Global.Infow("found device",
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

	d := &Device{
		control:  control,
		device:   device,
		callback: callback,
	}
	d.heading.Store(math.NaN())

	d.closeCh = make(chan struct{})
	go func() {
		ticker := time.NewTicker(time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-d.closeCh:
				return
			case <-ticker.C:
			}

			if callback.PacketAvailable() {
				packet := callback.GetNextPacket()
				if packet.ContainsOrientation() {
					euler := packet.OrientationEuler()
					if yaw := euler.Yaw(); !math.IsNaN(yaw) {
						d.heading.Store(yaw)
					}
				}
			}

		}
	}()
	return d, nil
}

func (d *Device) StartCalibration(ctx context.Context) error {
	return nil
}

func (d *Device) StopCalibration(ctx context.Context) error {
	return nil
}

func (d *Device) Readings(ctx context.Context) ([]interface{}, error) {
	heading, err := d.Heading(ctx)
	if err != nil {
		return nil, err
	}
	return []interface{}{heading}, nil
}

func (d *Device) Desc() sensor.DeviceDescription {
	return sensor.DeviceDescription{
		Type: compass.DeviceType,
	}
}

func (d *Device) Heading(ctx context.Context) (float64, error) {
	return d.heading.Load().(float64), nil
}

func (d *Device) Close(ctx context.Context) error {
	d.closeOnce.Do(func() {
		close(d.closeCh)
		defer d.control.Destruct()
		defer gen.DeleteCallbackHandler(d.callback)
	})
	return nil
}
