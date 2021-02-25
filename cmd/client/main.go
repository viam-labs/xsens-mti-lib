package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"time"

	mtiws "github.com/viamrobotics/mti/ws"

	"github.com/edaniels/golog"
)

func main() {
	port := 4444
	if flag.NArg() >= 1 {
		portParsed, err := strconv.ParseInt(flag.Arg(0), 10, 32)
		if err != nil {
			golog.Global.Fatal(err)
		}
		port = int(portParsed)
	}

	dev, err := mtiws.NewDevice(context.Background(), fmt.Sprintf("ws://localhost:%d", port))
	if err != nil {
		golog.Global.Fatal(err)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

READ:
	for {
		time.Sleep(time.Second)
		select {
		case <-sig:
			break READ
		default:
		}

		heading, err := dev.Heading(context.Background())
		if err != nil {
			if errors.Is(err, io.EOF) {
				break READ
			}
			golog.Global.Fatal(err)
		}
		golog.Global.Infow("heading", "data", heading)
	}
}
