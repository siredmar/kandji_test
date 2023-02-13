package session

import (
	"encoding/json"
	"fmt"
)

// SSHFlavor distinguishes between different kinds of SSH configurations on the target,
// which require some adaptations along the complete session (e.g. in path of special files).
type SSHFlavor uint

const (
	// SSHFlavorDropbearGridOS marks that the target uses Dropbear with GridOS-specific tweaks to file locations.
	SSHFlavorDropbearGridOS = iota
	// SSHFlavorOpenSSH marks a standard OpenSSH installation.
	SSHFlavorOpenSSH
)

// String implements fmt.Stringer.
func (s SSHFlavor) String() string {
	switch s {
	case SSHFlavorDropbearGridOS:
		return "dropbear-gridos"
	case SSHFlavorOpenSSH:
		return "openssh"
	}
	return "<invalid>"
}

// MarshalJSON implements json.Marshaler.
func (s SSHFlavor) MarshalJSON() ([]byte, error) {
	return []byte(s.String()), nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (s *SSHFlavor) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}
	switch str {
	case "dropbear-gridos":
		*s = SSHFlavorDropbearGridOS
	case "openssh":
		*s = SSHFlavorOpenSSH
	}
	return fmt.Errorf("invalid SSH flavor %q", str)
}

// KeysPath returns the directory where the authorized_keys files resides on the target.
func (s SSHFlavor) KeysPath() string {
	switch s {
	case SSHFlavorDropbearGridOS:
		return "/etc/dropbear"
	case SSHFlavorOpenSSH:
		// TODO this is actually dependent on the user we want to log in as. Stick with root-only for now.
		return "/root/.ssh"
	}
	return ""
}
