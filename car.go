package main

import (
	"fmt"

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
		fmt.Println("You are not authenticated...")
		return gatt.StatusUnexpectedError
	}

	// don't do anything if the state already matches the request
	if len(data) != 1 ||
		(c.isLocked && data[0] == LOCK_CAR) ||
		(!c.isLocked && data[0] == UNLOCK_CAR) {
		return gatt.StatusSuccess
	}

	// Pull the lever, Kronk!
	switch data[0] {
	case LOCK_CAR:
		c.Pin.Set()
		c.isLocked = true
		break
	case UNLOCK_CAR:
		c.Pin.Clear()
		c.isLocked = false
		break
	}

	return gatt.StatusSuccess
}
