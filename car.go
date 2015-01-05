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
		fmt.Println("Locking car")
		c.Lock()
		break
	case UNLOCK_CAR:
		fmt.Println("Unlocking car")
		c.Unlock()
		break
	}

	return gatt.StatusSuccess
}

func (c Car) Unlock() {
	if c.isLocked {
		c.Pin.Set()
		c.isLocked = false
	}
}

func (c Car) Lock() {
	if !c.isLocked {
		c.Pin.Clear()
		c.isLocked = true
	}
}
