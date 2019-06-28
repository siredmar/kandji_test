package printer

import (
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
	Labels            string `header:"Labels"`
}

func (o DeviceConsoleOutput) Map(d api.Device) DeviceConsoleOutput {
	o.ID = d.Metadata.ID
	o.Serialnumber = d.Spec.Serialnumber
	if d.Spec.MaintenanceWindow != nil {
		o.MaintenanceWindow = d.Spec.MaintenanceWindow.String()
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
		o.MaintenanceWindow = d.Spec.MaintenanceWindow.String()
	}
	if d.Spec.MACAddress != nil {
		o.MACAddress = *d.Spec.MACAddress
	}
	if d.Status.LastHeartbeat != nil {
		o.LastHeartbeat = d.Status.LastHeartbeat.String()
	}
	if d.Metadata.Labels != nil {
		o.Labels = SortedString(d.Metadata.Labels)
	}
	return o
}
