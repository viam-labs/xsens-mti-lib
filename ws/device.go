package mtiws

import (
	"context"
	"math"

	"github.com/edaniels/wsapi"
	"github.com/viamrobotics/robotcore/sensor/compass"
	"nhooyr.io/websocket"
)

const (
	CommandHeading = "heading"
)

type Device struct {
	conn *websocket.Conn
}

func NewDevice(ctx context.Context, address string) (compass.Device, error) {
	conn, _, err := websocket.Dial(ctx, address, nil)
	if err != nil {
		return nil, err
	}
	conn.SetReadLimit(10 * (1 << 24))

	return &Device{conn}, nil
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

func (d *Device) Heading(ctx context.Context) (float64, error) {
	if err := wsapi.WriteCommand(ctx, wsapi.NewCommand(CommandHeading), d.conn); err != nil {
		return math.NaN(), err
	}
	var heading float64
	err := wsapi.ReadJSONResponse(ctx, d.conn, &heading)
	return heading, err
}

func (d *Device) Close(ctx context.Context) error {
	defer d.conn.Close(websocket.StatusNormalClosure, "")
	return nil
}
