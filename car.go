package main

import (
	"time"

	"github.com/davecheney/gpio"
	"github.com/paypal/gatt"
)

type Car struct {
	Pin  gpio.Pin
	Auth *Auth
}

func NewCar(pin gpio.Pin, auth *Auth) *Car {
	return &Car{
		Pin:  pin,
		Auth: auth,
	}
}

func (c Car) HandleWrite(r gatt.Request, data []byte) (status byte) {
	if !c.Auth.IsAuthenticated() {
		return gatt.StatusUnexpectedError
	}

	go func() {
		c.Pin.Set()
		time.Sleep(100 * time.Millisecond)
		c.Pin.Clear()
	}()

	return gatt.StatusSuccess
}
