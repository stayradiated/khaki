package main

import (
	"log"
	"sync"
	"time"

	"github.com/paypal/gatt"
	"github.com/stayradiated/shifty"
)

// These bytes are sent over BLE
const UNLOCKED byte = 0x01
const NOTIFYING byte = 0x02

// Car represents a car
type Car struct {
	Auth *Auth

	mu          sync.Mutex
	Pin         shifty.Pin
	isUnlocked  bool
	isNotifying bool
	notifier    gatt.Notifier
}

// NewCar creates a new instance of the Car type
func NewCar(pin shifty.Pin, auth *Auth) *Car {
	return &Car{
		Pin:         pin,
		Auth:        auth,
		isUnlocked:  false,
		isNotifying: true,
	}
}

// HandleWrite will lock or unlock the car
func (c *Car) Write(data []byte) (status byte) {
	if !c.Auth.IsAuthenticated() {
		log.Println("You are not authenticated...")
		return gatt.StatusUnexpectedError
	}

	if len(data) < 1 {
		log.Println("Invalid data")
		return gatt.StatusUnexpectedError
	}

	// Pull the lever, Kronk!
	if data[0]&UNLOCKED == 0 {
		log.Println("--- Locking car")
		c.Lock()
	} else {
		log.Println("+++ Unlocking car")
		c.Unlock()
	}

	return gatt.StatusSuccess
}

// HandleNotify sends the current  status of the car to the central every two
// seconds
func (c *Car) Notify(n gatt.Notifier) {
	c.mu.Lock()
	c.notifier = n
	c.mu.Unlock()

	go func() {
		for !n.Done() {

			c.mu.Lock()
			isNotifying := c.isNotifying
			c.mu.Unlock()

			if isNotifying {
				n.Write(c.Status())
			}

			time.Sleep(10 * time.Second)
		}
	}()
}

// Unlock unlocks the car
func (c *Car) Unlock() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isUnlocked {
		go c.pressButton()
		c.isUnlocked = true
	}
}

// Lock locks the car
func (c *Car) Lock() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isUnlocked {
		go c.pressButton()
		c.isUnlocked = false
	}
}

// Press the button
func (c *Car) pressButton() {
	if c.Pin != nil {
		c.Pin.Set()
		time.Sleep(500 * time.Millisecond)
		c.Pin.Clear()
	}
}

func (c *Car) ToggleNotifications(status bool) {
	c.mu.Lock()
	c.isNotifying = status
	notifier := c.notifier
	c.mu.Unlock()

	if notifier == nil {
		log.Println("No connection")
		return
	}

	if status {
		log.Println("Enabling notifications")
	} else {
		log.Println("Stopping notifications")
	}

	notifier.Write(c.Status())
}

func (c *Car) Reset() {
	c.mu.Lock()
	c.notifier = nil
	c.mu.Unlock()
}

func (c *Car) Status() []byte {
	status := byte(0)

	c.mu.Lock()
	if c.isUnlocked {
		status |= UNLOCKED
	}
	if c.isNotifying {
		status |= NOTIFYING
	}
	c.mu.Unlock()

	log.Printf("%x", status)

	return []byte{status}
}
