package ssh

import (
	"github.com/google/uuid"
)

// Message types
const (
	CreateProcessMessageType = iota
	ExecuteCommandMessageType
	ProcessOutputMessageType
	ProcessCreatedMessageType
	ProcessTerminatedMessageType
	CreateFileMessageType
	WriteToFileMessageType
	ErrorMessageType
)

// GetUUID represents a function returning a v4 uuid
type GetUUID func() uuid.UUID

// RAWMessage represents a ssh message
type RAWMessage struct {
	MessageType
}

// MessageType represents a ssh message
type MessageType struct {
	Type int `json:"type"`
}

// CreateProcessMessage represents a ssh message of type create process
type CreateProcessMessage struct {
	MessageType
	ID      string `json:"id"`
	Command []byte `json:"command"`
}

// NewCreateProcessMessage creates a new ssh message of type create process
func NewCreateProcessMessage(i string, c []byte) CreateProcessMessage {
	msg := CreateProcessMessage{}
	msg.Type = CreateProcessMessageType
	msg.ID = i
	msg.Command = c

	return msg
}

// CreateFileMessage represents a ssh message of type create file
// Inverse can be used to trigger a backcopy from device->client
// if set to true, the ssh-agent on the device will start reading
// the file and send it's content to the client.
type CreateFileMessage struct {
	MessageType
	ID       string `json:"id"`
	Filename string `json:"filename"`
	Inverse  bool   `json:"inverse"`
}

// NewCreateFileMessage creates a new ssh message of type create process
func NewCreateFileMessage(i, f string, in bool) CreateFileMessage {
	msg := CreateFileMessage{}
	msg.Type = CreateFileMessageType
	msg.ID = i
	msg.Filename = f
	msg.Inverse = in

	return msg
}

// ExecuteCommandMessage represents a ssh message of type execute command
type ExecuteCommandMessage struct {
	MessageType
	ID      string `json:"id"`
	Command []byte `json:"command"`
}

// NewExecuteCommandMessage creates a new ssh message of type execute command
func NewExecuteCommandMessage(i string, c []byte) ExecuteCommandMessage {
	msg := ExecuteCommandMessage{}
	msg.Type = ExecuteCommandMessageType
	msg.ID = i
	msg.Command = c

	return msg
}

// WriteToFileMessage represents a ssh message of type write to file
type WriteToFileMessage struct {
	MessageType
	ID       string `json:"id"`
	Filename string `json:"filename"`
	Content  []byte `json:"content"`
	EOF      bool   `json:"eof"`
}

// NewWriteToFileMessage creates a new ssh message of type execute command
func NewWriteToFileMessage(i, f string, c []byte, e bool) WriteToFileMessage {
	msg := WriteToFileMessage{}
	msg.Type = WriteToFileMessageType
	msg.ID = i
	msg.Filename = f
	msg.Content = c
	msg.EOF = e

	return msg
}

// ProcessCreatedMessage represents a ssh message of type process create
type ProcessCreatedMessage struct {
	MessageType
	ID string `json:"id"`
}

// NewProcessCreatedMessage creates a new ssh message of type process created
func NewProcessCreatedMessage(i string) CreateProcessMessage {
	msg := CreateProcessMessage{}
	msg.Type = ProcessCreatedMessageType
	msg.ID = i

	return msg
}

// ProcessOutputMessage represents a ssh message of type process output
type ProcessOutputMessage struct {
	MessageType
	ID   string `json:"id"`
	Data []byte `json:"data"`
}

// NewProcessOutputMessage creates a new ssh message of type process output
func NewProcessOutputMessage(i string, d []byte) ProcessOutputMessage {
	msg := ProcessOutputMessage{}
	msg.Type = ProcessOutputMessageType
	msg.ID = i
	msg.Data = d

	return msg
}

// ProcessTerminatedMessage represents a ssh message of type process terminated
type ProcessTerminatedMessage struct {
	MessageType
	ID     string `json:"id"`
	Reason []byte `json:"reason"`
}

// NewProcessTerminatedMessage creates a new ssh message of type process terminated
func NewProcessTerminatedMessage(i string, r []byte) ProcessTerminatedMessage {
	msg := ProcessTerminatedMessage{}
	msg.Type = ProcessTerminatedMessageType
	msg.ID = i
	msg.Reason = r

	return msg
}

// ErrorMessage represents a ssh message of type error
type ErrorMessage struct {
	MessageType
	ErrorMsg string `json:"error"`
}

// NewErrorMessage creates a new ssh message of type error
func NewErrorMessage(m string) ErrorMessage {
	msg := ErrorMessage{}
	msg.Type = ErrorMessageType
	msg.ErrorMsg = m

	return msg
}
