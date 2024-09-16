package spinner

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	enabled   = true
	frames    = []string{`⠋`, `⠙`, `⠹`, `⠸`, `⠼`, `⠴`, `⠦`, `⠧`, `⠇`, `⠏`}
	framesNum = 10
	interval  = 60 * time.Millisecond
)

// Spinner displays a waiting spinner in the terminal
type Spinner struct {
	done    chan bool
	enabled bool
	out     *os.File
	sigs    chan os.Signal
	ticker  *time.Ticker
}

// New Spinner
func New(msg string) *Spinner {
	if !enabled {
		return &Spinner{
			enabled: enabled,
		}
	}
	s := &Spinner{
		done:    make(chan bool, 1),
		enabled: enabled,
		out:     os.Stderr,
		sigs:    make(chan os.Signal, 1),
		ticker:  time.NewTicker(interval),
	}
	signal.Notify(s.sigs, syscall.SIGINT, syscall.SIGTERM)
	fmt.Fprintf(s.out, "      %v", msg)
	s.start()
	return s
}

// Disable spinners globally
func Disable() {
	enabled = false
}

// Write just writes out a message for the spinner
func (s *Spinner) Write(t string) {
	if !s.enabled {
		return
	}
	fmt.Fprintf(s.out, "\r"+t+"\n")
}

// Ok indicates success
func (s *Spinner) Ok() {
	if !s.enabled {
		return
	}
	s.stop()
	fmt.Fprintf(s.out, "\r \033[32mOK\033[39m\n")
}

// Warn indicates error
func (s *Spinner) Warn() {
	if !s.enabled {
		return
	}
	s.stop()
	fmt.Fprintf(s.out, "\r \033[33mWARN\033[39m\n")
}

// Fail indicates failure
func (s *Spinner) Fail() {
	if !s.enabled {
		return
	}
	s.stop()
	fmt.Fprintf(s.out, "\r \033[31mFAIL\033[39m\n")
}

func (s *Spinner) start() {
	if !s.enabled {
		return
	}
	var i int
	// disable cursor
	fmt.Fprintf(s.out, "\033[?25l")
	go func() {
		for {
			select {
			case <-s.sigs:
				s.stop()
				os.Exit(0)
			case <-s.done:
				return
			case <-s.ticker.C:
				fmt.Fprintf(s.out, "\r %v", frames[i%framesNum])
				i++
			}
		}
	}()
}

func (s *Spinner) stop() {
	if !s.enabled {
		return
	}
	// enable cursor
	fmt.Fprintf(s.out, "\033[?25h")
	s.ticker.Stop()
	s.done <- true
}
