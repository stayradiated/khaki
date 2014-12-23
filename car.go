package main

import (
	"time"

	"github.com/davecheney/gpio"
	"github.com/paypal/gatt"
)

type Car struct {
	Pin gpio.Pin
}

func (c Car) HandleWrite(r gatt.Request, data []byte) (status byte) {
	if !authed {
		return gatt.StatusUnexpectedError
	}

	go func() {
		c.Pin.Set()
		time.Sleep(100 * time.Millisecond)
		c.Pin.Clear()
	}()

	return gatt.StatusSuccess
}
