package std

import (
	"os"
)

// Std stream wraps the standard input and output streams
type Std struct{}

// New returns a Std stream.
func New() *Std {
	return &Std{}
}

// Read from stdin into buffer.
func (std *Std) Read(b []byte) (int, error) {
	return os.Stdin.Read(b)
}

// Write buffer to stdout.
func (std *Std) Write(b []byte) (int, error) {
	return os.Stdout.Write(b)
}
