package main

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// Spinner is a simple spinner
//
// Example:
//
//	spinner := NewSpinner(100*time.Millisecond,
//		[]string{ "⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏", })
//	spinner.Start()
//	time.Sleep(5 * time.Second)
//	spinner.Stop()
//	fmt.Println("Done.")
type Spinner struct {
	running  bool
	interval time.Duration
	frames   []string
	cursor   int

	mu sync.Mutex
}

// NewSpinner creates a new spinner
func NewSpinner(interval time.Duration, frames []string) *Spinner {
	return &Spinner{
		running:  false,
		interval: interval,
		frames:   frames,
		cursor:   0,
	}
}

// Start starts spinner
func (s *Spinner) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return
	}

	s.running = true
	go func() {
		for s.running {
			fmt.Printf("\r%s", s.frames[s.cursor])
			time.Sleep(s.interval)
			s.cursor = (s.cursor + 1) % len(s.frames)
		}
	}()
}

// Stop stops spinner
func (s *Spinner) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	s.running = false
	fmt.Printf("\r%s\r", strings.Repeat(" ", len(s.frames[0])))
}
