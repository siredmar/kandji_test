package printer

import (
	"fmt"

	api "github.com/grid-x/gxctl/pkg/api"
)

type DeviceConsoleOutput struct {
	ID                string `header:"ID"`
	Serialnumber      string `header:"Serialnumber"`
	MaintenanceWindow string `header:"Maintenance Window"`
	MACAddress        string `header:"MAC address"`
}

type DeviceConsoleOutputWide struct {
	ID                string `header:"ID"`
	Serialnumber      string `header:"Serialnumber"`
	MaintenanceWindow string `header:"Maintenance Window"`
	MACAddress        string `header:"MAC address"`
	LastHeartbeat     string `header:"Last heartbeat"`
	PublicKey         string `header:"Pubkey"`
	Labels            string `header:"Labels"`
}

func (o DeviceConsoleOutput) Map(d api.Device) DeviceConsoleOutput {
	o.ID = d.Metadata.ID
	o.Serialnumber = d.Spec.Serialnumber
	if d.Spec.MaintenanceWindow != nil {
		o.MaintenanceWindow = *d.Spec.MaintenanceWindow
	}
	if d.Spec.MACAddress != nil {
		o.MACAddress = *d.Spec.MACAddress
	}
	return o
}

func (o DeviceConsoleOutputWide) Map(d api.Device) DeviceConsoleOutputWide {
	o.ID = d.Metadata.ID
	o.Serialnumber = d.Spec.Serialnumber
	if d.Spec.MaintenanceWindow != nil {
		o.MaintenanceWindow = *d.Spec.MaintenanceWindow
	}
	if d.Spec.MACAddress != nil {
		o.MACAddress = *d.Spec.MACAddress
	}
	if d.Status.LastHeartbeat != "" {
		o.LastHeartbeat = d.Status.LastHeartbeat
	}
	if d.Spec.PublicKey != nil {
		o.PublicKey = *d.Spec.PublicKey
		if len(o.PublicKey) > 50 {
			o.PublicKey = o.PublicKey[0:50] + "..."
		}
	}
	if d.Metadata.Labels != nil {
		var s string
		for key, value := range d.Metadata.Labels {
			s += fmt.Sprintf("%s:%s\n", key, value)
		}

		o.Labels = s[:len(s)-1]
	}
	return o
}
