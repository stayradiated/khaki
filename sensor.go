package main

import "github.com/davecheney/gpio"

type Sensor struct {
	Pin gpio.Pin
}

func (s *Sensor) Watch(handler func()) {
	s.Pin.BeginWatch(gpio.EdgeBoth, handler)
}
