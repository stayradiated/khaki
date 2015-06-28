package main

import (
	"sync"
	"time"

	"github.com/stayradiated/shifty"
)

type StatusLED struct {
	Pin   shifty.Pin
	Blink time.Duration

	mu     sync.Mutex
	status bool
}

func NewStatusLED(pin shifty.Pin, blink time.Duration, status bool) *StatusLED {
	led := &StatusLED{Pin: pin, Blink: blink}
	led.Update(status)
	return led
}

// Update sets the status of the LED
func (s *StatusLED) Update(status bool) {
	s.mu.Lock()
	s.status = status
	s.mu.Unlock()

	if status {
		s.openPin()
		s.startBlinking()
	} else {
		s.closePin()
	}
}

func (s *StatusLED) startBlinking() {
	if s.Blink > 0 {
		ticker := time.NewTicker(s.Blink)

		go func() {
			for _ = range ticker.C {

				if s.status == false {
					ticker.Stop()
					break
				}

				s.closePin()
				time.Sleep(time.Millisecond * 200)
				s.openPin()
			}
		}()
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
