package main

import (
	"os"
	"os/signal"

	"github.com/davecheney/gpio"
	"github.com/davecheney/gpio/rpi"
)

// OpenGPIOPin opens up a GPIO pin
func OpenGPIOPin() (gpio.Pin, error) {

	// use GPIO25 pin
	pin, err := gpio.OpenPin(rpi.GPIO25, gpio.ModeOutput)
	if err != nil {
		return nil, err
	}

	// turn the led off at exit
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
