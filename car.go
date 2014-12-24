package main

import (
	"time"

	"github.com/davecheney/gpio"
	"github.com/paypal/gatt"
)

const UNLOCK_CAR = 1
const LOCK_CAR = 2

type Car struct {
	isLocked bool
	Pin      gpio.Pin
	Auth     *Auth
}

func NewCar(pin gpio.Pin, auth *Auth) *Car {
	return &Car{
		isLocked: false,
		Pin:      pin,
		Auth:     auth,
	}
}

func (c Car) HandleWrite(r gatt.Request, data []byte) (status byte) {
	if !c.Auth.IsAuthenticated() {
		return gatt.StatusUnexpectedError
	}

	// don't do anything if the state already matches the request
	if len(data) != 1 ||
		(c.isLocked && data[0] == LOCK_CAR) ||
		(!c.isLocked && data[0] == UNLOCK_CAR) {
		return gatt.StatusSuccess
	}

	// pull the level kronk!
	go func() {
		c.Pin.Set()
		time.Sleep(100 * time.Millisecond)
		c.Pin.Clear()
	}()

	return gatt.StatusSuccess
}
