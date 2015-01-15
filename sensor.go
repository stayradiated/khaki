package main

import "github.com/davecheney/gpio"

type Sensor struct {
	Pin          gpio.Pin
	HandleChange func(bool)
}

func (s *Sensor) Init() {
	s.HandleChange(s.Pin.Get())
	s.Pin.BeginWatch(gpio.EdgeBoth, func() {
		s.HandleChange(s.Pin.Get())
	})
}
