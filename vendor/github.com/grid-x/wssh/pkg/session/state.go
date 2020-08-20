package session

// State in which a Session can be in
type State int

const (
	// StateInit is the state a Session is in after creation
	StateInit State = iota
	// StateWaitDevice indicates waiting for a Device
	StateWaitDevice
	// StateReady indicates working clients
	StateReady
	// StateClosing indicates a closing Session
	StateClosing
	// StateCloseAgent indicates the Agent is closing
	StateCloseAgent
	// StateCloseDevice indicates the Device is closing
	StateCloseDevice
	// StateFin indicates the Session is finished
	StateFin
)

func (s State) String() string {
	return [...]string{
		"Init",
		"WaitDevice",
		"Ready",
		"Closing",
		"CloseAgent",
		"CloseDevice",
		"Fin",
	}[s]
}
