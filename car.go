package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/davecheney/gpio"
	"github.com/paypal/gatt"
)

// These bytes are sent over BLE
const UNLOCKED byte = 0x01
const LOCKED byte = 0x02

// Car represents a car
type Car struct {
	Auth *Auth

	mu     sync.Mutex
	Status byte
	Pin    gpio.Pin
}

// NewCar creates a new instance of the Car type
func NewCar(pin gpio.Pin, auth *Auth) *Car {
	return &Car{
		Status: LOCKED,
		Pin:    pin,
		Auth:   auth,
	}
}

// HandleWrite will lock or unlock the car
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
	switch data[0] {
	case LOCKED:
		fmt.Println("--- Locking car")
		c.Lock()
		break
	case UNLOCKED:
		fmt.Println("+++ Unlocking car")
		c.Unlock()
		break
	}

	return gatt.StatusSuccess
}

// HandleRead reports the current status of the car
func (c *Car) HandleRead(resp gatt.ReadResponseWriter, req *gatt.ReadRequest) {
	c.mu.Lock()
	status := c.Status
	c.mu.Unlock()

	resp.Write([]byte{status})
}

// HandleNotify sends the current  status of the car to the central every two
// seconds
func (c *Car) HandleNotify(r gatt.Request, n gatt.Notifier) {
	go func() {
		for !n.Done() {
			c.mu.Lock()
			status := c.Status
			c.mu.Unlock()

			n.Write([]byte{status})
			time.Sleep(10 * time.Second)
		}
	}()
}

// Unlock unlocks the car
func (c *Car) Unlock() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Status == LOCKED {
		if c.Pin != nil {
			c.Pin.Set()
		}
		c.Status = UNLOCKED
	}
}

// Lock locks the car
func (c *Car) Lock() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Status == UNLOCKED {
		if c.Pin != nil {
			c.Pin.Clear()
		}
		c.Status = LOCKED
	}
}
