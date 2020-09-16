package tunnel

// State in which a Tunnel can be in
type State int

const (
	// StateInit is the state a Tunnel is in after creation
	StateInit State = iota
	// StateReady indicates a working connection
	StateReady
	// StateClosing indicates the tunnel is closing
	StateClosing
	// StateFin indicates the Tunnel is finished
	StateFin
)

func (s State) String() string {
	return [...]string{"Init", "Ready", "Closing", "Fin"}[s]
}
