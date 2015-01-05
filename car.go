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

	if len(data) < 1 {
		fmt.Println("Invalid data")
		return gatt.StatusUnexpectedError
	}

	// Pull the lever, Kronk!
	switch data[0] {
	case LOCK_CAR:
		c.Lock()
		break
	case UNLOCK_CAR:
		c.Unlock()
		break
		break
	}

	return gatt.StatusSuccess
}

func (c Car) Unlock() {
	c.Pin.Set()
	c.isLocked = true
}

func (c Car) Lock() {
	c.Pin.Clear()
	c.isLocked = false
}
