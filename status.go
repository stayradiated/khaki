package main

import "github.com/davecheney/gpio"

type Status struct {
	Connected bool
	Pin       gpio.Pin
}

func (s *Status) Update(connected bool) {
	s.Connected = connected
	if s.Connected {
		s.openPin()
	} else {
		s.closePin()
	}
}

func (s Status) openPin() {
	if s.Pin != nil {
		s.Pin.Set()
	}
}

func (s Status) closePin() {
	if s.Pin != nil {
		s.Pin.Clear()
	}
}
