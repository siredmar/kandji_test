package printer

import (
	api "github.com/grid-x/gxctl/pkg/api"
)

type DeviceConsoleOutput struct {
	ID                string `header:"ID"`
	Serialnumber      string `header:"Serialnumber"`
	MaintenanceWindow string `header:"Maintenance Window"`
	MACAddress        string `header:"MAC address"`
	LastHeartbeat     string `header:"Last heartbeat"`
}

type DeviceConsoleOutputWide struct {
	ID                string `header:"ID"`
	Serialnumber      string `header:"Serialnumber"`
	MaintenanceWindow string `header:"Maintenance Window"`
	MACAddress        string `header:"MAC address"`
	LastHeartbeat     string `header:"Last heartbeat"`
	FirstSeen         string `header:"First seen"`
	SuperviorVersion  string `header:"Supervisor"`
	OSVersion         string `header:"OS"`
	Labels            string `header:"Labels"`
	Annotations       string `header:"Annotations"`
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
	if d.Status.LastHeartbeat != nil {
		o.LastHeartbeat = d.Status.LastHeartbeat.Format("02.01.2006 15:04:05")
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
		o.LastHeartbeat = d.Status.LastHeartbeat.Format("02.01.2006 15:04:05")
	}
	if d.Status.FirstSeen != nil {
		o.FirstSeen = d.Status.FirstSeen.Format("02.01.2006 15:04:05")
	}
	if d.Metadata.Labels != nil {
		o.Labels = SortedString(d.Metadata.Labels)
	}
	if d.Metadata.Annotations != nil {
		o.Annotations = SortedString(d.Metadata.Annotations)
	}
	if d.Status.Info != nil {
		if d.Status.Info.SupervisorVersion != nil {
			o.SuperviorVersion = *d.Status.Info.SupervisorVersion
		}
		if d.Status.Info.OSVersion != nil {
			o.OSVersion = *d.Status.Info.OSVersion
		}
	}
	return o
}
