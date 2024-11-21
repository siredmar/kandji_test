package printer

import (
	"time"

	api "github.com/grid-x/gxctl/pkg/api"
	units "github.com/grid-x/gxctl/pkg/printer/units"
)

type DeviceConsoleOutput struct {
	ID               string `header:"ID"`
	Serialnumber     string `header:"Serialnumber"`
	LastHeartbeat    string `header:"Last heartbeat"`
	SuperviorVersion string `header:"Supervisor"`
	OSVersion        string `header:"OS"`
}

type DeviceConsoleOutputWide struct {
	ID               string `header:"ID"`
	Serialnumber     string `header:"Serialnumber"`
	LastHeartbeat    string `header:"Last heartbeat"`
	FirstSeen        string `header:"First seen"`
	SuperviorVersion string `header:"Supervisor"`
	OSVersion        string `header:"OS"`
	Labels           string `header:"Labels"`
	Annotations      string `header:"Annotations"`
	PublicIP         string `header:"Public IPv4"`
}

func (o DeviceConsoleOutput) Map(d api.Device) DeviceConsoleOutput {
	offlineIndicator := "* "
	if d.IsOnline() {
		offlineIndicator = " "
	}
	o.ID = offlineIndicator + d.Metadata.ID
	o.Serialnumber = d.Spec.Serialnumber
	if d.Status.LastHeartbeat != nil {
		o.LastHeartbeat = units.HumanDuration(time.Since(d.Status.LastHeartbeat.Time)) + " ago"
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

func (o DeviceConsoleOutputWide) Map(d api.Device) DeviceConsoleOutputWide {
	o.ID = d.Metadata.ID
	o.Serialnumber = d.Spec.Serialnumber
	if d.Status.LastHeartbeat != nil {
		o.LastHeartbeat = units.HumanDuration(time.Since(d.Status.LastHeartbeat.Time)) + " ago"
	}
	if d.Spec.FirstSeen != nil {
		o.FirstSeen = d.Spec.FirstSeen.Format("02.01.2006 15:04:05")
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
		if d.Status.Info.PublicIP != nil {
			o.PublicIP = *d.Status.Info.PublicIP
		}
	}
	return o
}
