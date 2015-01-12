package main

import (
	"sync"

	"github.com/davecheney/gpio"
)

type StatusLED struct {
	Pin gpio.Pin

	mu     sync.Mutex
	status bool
}

// Update sets the status of the LED
func (s *StatusLED) Update(status bool) {
	s.mu.Lock()
	s.status = status
	s.mu.Unlock()

	if status {
		s.openPin()
	} else {
		s.closePin()
	}
}

// openPin turns the pin on
func (s StatusLED) openPin() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Pin != nil {
		s.Pin.Set()
	}
}

// closePin turns the pin off
func (s StatusLED) closePin() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Pin != nil {
		s.Pin.Clear()
	}
}
