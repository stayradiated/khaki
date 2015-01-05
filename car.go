package main

import (
	"fmt"
	"sync"

	"github.com/davecheney/gpio"
	"github.com/paypal/gatt"
)

const UNLOCK_CAR = 1
const LOCK_CAR = 2

var mu = &sync.Mutex{}

type Car struct {
	isLocked bool
	Pin      gpio.Pin
	Auth     *Auth
}

func NewCar(pin gpio.Pin, auth *Auth) *Car {
	return &Car{
		isLocked: true,
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

	if data[0] == LOCK_CAR {
		fmt.Println("--- Locking car")
		c.Lock()
	} else if data[0] == UNLOCK_CAR {
		fmt.Println("+++ Unlocking car")
		c.Unlock()
	}

	return gatt.StatusSuccess
}

func (c Car) Unlock() {
	mu.Lock()
	// if c.isLocked == true {
	fmt.Println("Setting LED")
	c.Pin.Set()
	c.isLocked = false
	// }
	mu.Unlock()
}

func (c Car) Lock() {
	mu.Lock()
	// if c.isLocked == false {
	fmt.Println("Clearing LED")
	c.Pin.Clear()
	c.isLocked = true
	// }
	mu.Unlock()
}
