package printer

import (
	api "github.com/grid-x/gxctl/pkg/api"
)

type DeviceConsoleOutput struct {
	ID            string `header:"ID"`
	Serialnumber  string `header:"Serialnumber"`
	MACAddress    string `header:"MAC address"`
	LastHeartbeat string `header:"Last heartbeat"`
	PublicKey     string `header:"Pubkey"`
}

func (o DeviceConsoleOutput) Map(d api.Device) DeviceConsoleOutput {
	o.ID = d.ID
	o.Serialnumber = d.Spec.Serialnumber
	if d.Spec.MACAddress != nil {
		o.MACAddress = *d.Spec.MACAddress
	}
	if d.Status.LastHeartbeat != nil {
		o.LastHeartbeat = *d.Status.LastHeartbeat
	}
	if d.Spec.PublicKey != nil {
		o.PublicKey = *d.Spec.PublicKey
		if len(o.PublicKey) > 50 {
			o.PublicKey = o.PublicKey[0:50] + "..."
		}
	}
	return o
}
