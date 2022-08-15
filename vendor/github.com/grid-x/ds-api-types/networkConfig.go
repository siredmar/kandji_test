package types

import (
	"errors"
	"fmt"
	"net"
)

// NetworkConfig represents the complete desired state of network configuration
// of a device.
type NetworkConfig map[string]NetworkInterfaceSpec

// Validate returns an error if the given network configuration is semantically invalid
func (n NetworkConfig) Validate() error {
	if len(n) == 0 {
		return nil
	}
	numGateways := 0
	for intf, spec := range n {
		if err := spec.Validate(); err != nil {
			return fmt.Errorf("invalid config for %q: %w", intf, err)
		}
		if spec.Gateway != nil {
			numGateways++
		}
	}
	if numGateways == 0 {
		return errors.New("no interface with configured gateway")
	}
	return nil
}

// NetworkInterfaceSpec represents the desired state of a network interface's configuration
type NetworkInterfaceSpec struct {
	// IPv4 address to be assigned to this interface, or the special string "dhcp" to use DHCP.
	Address string `json:"address"`
	// If non-empty, a default route via this interface is added. If "address" is "dhcp", this value is ignored and the
	// gateway provided via DHCP is used. Otherwise, this must be an IPv4 address that is used as the gateway for the
	// default route.
	Gateway *string `json:"gateway,omitempty"`
	// Specifies the "route metric" for routes through this interface. Effectively, this is a priority setting for this
	// interface, with lower values representing a higher priority. If unset, default values based on the interface type
	// are chosen, with wired connections getting a lower metric than cellular ones.
	Metric *int64 `json:"metric,omitempty"`
	// If true, this interface is only activated if all other interfaces with a configured Gateway and a lower Metric
	// are unavailable for Internet connectivity, and is deactivated once such an interface becomes available. Otherwise
	// (the default), this interface will be held constantly activated as far as possible.
	OnDemand bool `json:"onDemand,omitempty"`
	// If "address" is not "dhcp", this is used as the subnet mask for this interface's route. Otherwise this value is ignored.
	Subnet *string `json:"subnet,omitempty"`
	// Settings specifically for mobile broadband devices.
	ModemConfig *ModemConfig `json:"modemConfig,omitempty"`
}

// ModemConfig specifies the configuration of the mobile broadband modem, if any.
type ModemConfig struct {
	// GPRS Access Point Name to use for establishing a data connection.
	APN string `json:"apn"`
	// Username to authenticate with the network, if required.
	Username *string `json:"username,omitempty"`
	// Password to authenticate with the network, if required.
	Password *string `json:"password,omitempty"`
	// PIN to unlock the SIM card, if required.
	PIN *string `json:"pin,omitempty"`
	// Disable roaming to other networks. Default is false, so roaming is allowed.
	DisableRoaming bool `json:"disableRoaming,omitempty"`
}

// Validate  returns an error if the given configuration for an interface is semantically invalid
func (spec NetworkInterfaceSpec) Validate() error {
	if spec.Address == "dhcp" {
		return nil
	}

	parsed := net.ParseIP(spec.Address)
	if parsed == nil || parsed.To4() == nil {
		return errors.New("invalid 'address'")
	}
	if spec.Subnet == nil {
		return errors.New("missing 'subnet'")
	}
	subnet := net.ParseIP(*spec.Subnet)
	if subnet == nil || subnet.To4() == nil {
		return errors.New("invalid 'subnet'")
	}
	mask := net.IPMask(subnet.To4())
	// If the mask is not in the canonical form--ones followed by zeros--then Size returns 0, 0.
	ones, _ := mask.Size()
	if ones == 0 {
		return errors.New("invalid 'subnet'")
	}

	if spec.Gateway != nil {
		gateway := net.ParseIP(*spec.Gateway)
		if gateway == nil || gateway.To4() == nil {
			return errors.New("invalid 'gateway'")
		}

		net := net.IPNet{
			IP:   gateway,
			Mask: mask,
		}
		if !net.Contains(parsed) {
			return errors.New("'address' and 'gateway' not in the same network")
		}
	}
	return nil
}
