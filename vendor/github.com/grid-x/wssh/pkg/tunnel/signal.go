package tunnel

// Signal is used by the Session to communicate with a Tunnel
type Signal int

const (
	// SignalOpen indicates an open Tunnel
	SignalOpen Signal = iota
	// SignalClose indicates an closing or closed Tunnel
	SignalClose
)

func (s Signal) String() string {
    return [...]string{"Open", "Close"}[s]
}
