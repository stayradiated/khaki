package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/davecheney/gpio"
	"github.com/paypal/gatt"
)

const UNLOCK_CAR = 1
const LOCK_CAR = 2

type Car struct {
	sync.Mutex

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

func (c *Car) HandleWrite(r gatt.Request, data []byte) (status byte) {
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
		c.Close()
	} else if data[0] == UNLOCK_CAR {
		fmt.Println("+++ Unlocking car")
		c.Open()
	}

	return gatt.StatusSuccess
}

func (c *Car) HandleNotify(r gatt.Request, n gatt.Notifier) {
	go func() {
		bytes := []byte{0}
		for !n.Done() {
			n.Write(bytes)
			bytes[0]++
			time.Sleep(2 * time.Second)
		}
	}()
}

func (c *Car) Open() {
	c.Lock()
	if c.isLocked == true {
		if c.Pin != nil {
			c.Pin.Set()
		}
		c.isLocked = false
	}
	c.Unlock()
}

func (c *Car) Close() {
	c.Lock()
	if c.isLocked == false {
		if c.Pin != nil {
			c.Pin.Clear()
		}
		c.isLocked = true
	}
	c.Unlock()
}
