package main

import (
	"log"

	"github.com/davecheney/gpio"
)

type Sensor struct {
	Pin gpio.Pin
}

func (s *Sensor) Watch() {
	s.Pin.BeginWatch(gpio.EdgeBoth, s.handleChange)
}

func (s *Sensor) handleChange() {
	log.Println(s.Pin.Get())
}
