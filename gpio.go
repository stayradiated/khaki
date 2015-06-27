package main

import (
	"os"
	"os/signal"

	"github.com/davecheney/gpio"
	"github.com/davecheney/gpio/rpi"
)

// OpenPinForOutput opens up a GPIO pin for output
func OpenPinForOutput(pinId int) (gpio.Pin, error) {

	// open pin
	pin, err := rpi.OpenPin(pinId, gpio.ModeOutput)
	if err != nil {
		return nil, err
	}

	// turn the pin off when we exit
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			pin.Clear()
			pin.Close()
			os.Exit(0)
		}
	}()

	return pin, nil
}

// MustOpenPinForOutput opens up a GPIO pin for output
// Will panic if pin can't be opened
func MustOpenPinForOutput(pinId int) gpio.Pin {
	pin, err := OpenPinForOutput(pinId)
	if err != nil {
		panic(err)
	}
	return pin
}

// OpenPinForInput opens up a GPIO pin for input
func OpenPinForInput(pinId int) (gpio.Pin, error) {

	// open pin
	pin, err := rpi.OpenPin(pinId, gpio.ModeInput)
	if err != nil {
		return nil, err
	}

	// close the pin at exit
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			pin.Close()
			os.Exit(0)
		}
	}()

	return pin, nil
}
