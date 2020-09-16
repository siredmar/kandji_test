package session

// Signal is used by the server to communicate with a session
type Signal int

const (
	// SignalSuccess indicates successful session setup
	SignalSuccess Signal = iota
	// SignalFailure indicates unsuccessful session setup
	SignalFailure
)

func (s Signal) String() string {
	return [...]string{"Success", "Failure"}[s]
}
