package api

import (
	"github.com/google/go-cmp/cmp"

	applicationsApi "github.com/grid-x/ds-api-types/management/2018-11-27/application"
	maintenanceApi "github.com/grid-x/ds-api-types/management/2018-12-04/maintenance"
	dockerConfigApi "github.com/grid-x/ds-api-types/management/2019-04-01/dockerconfigs"
	deviceDockerConfigApi "github.com/grid-x/ds-api-types/management/2019-05-14/devicedockerconfigs"
	devicesApi "github.com/grid-x/ds-api-types/management/2019-06-13/device"
	deploymentsApi "github.com/grid-x/ds-api-types/management/2019-08-17/deployments"
	podsApi "github.com/grid-x/ds-api-types/management/2019-08-17/pod"
)

type FullObjectMeta struct {
	Meta struct {
		Id   string `json:"id,omitempty"`
		Name string `json:"name,omitempty"`
	} `json:"metadata"`
}

type Device devicesApi.Device

type Devices struct {
	Devices []Device `json:"devices"`
}

type PatchDevice devicesApi.UpdateRequest

type Pod podsApi.Pod

type Pods struct {
	Pods []Pod `json:"pods"`
}

type Deployment deploymentsApi.Deployment

type CreateDeployment deploymentsApi.CreateRequest

type PatchDeployment deploymentsApi.UpdateRequest

type Deployments struct {
	Deployments []Deployment `json:"deployments"`
}

type Application applicationsApi.Application

type Applications struct {
	Applications []Application `json:"applications"`
}

type CreateApplication applicationsApi.CreateRequest

type MaintenanceTask maintenanceApi.Task

type MaintenanceTasks struct {
	MaintenanceTasks []MaintenanceTask `json:"maintenanceTasks"`
}

type CreateMaintenanceTask maintenanceApi.CreateRequest

type DockerConfig dockerConfigApi.DockerConfig

type DockerConfigs struct {
	DockerConfigs []DockerConfig `json:"dockerConfigs"`
}

type CreateDockerConfig dockerConfigApi.CreateRequest

type DeviceDockerConfig deviceDockerConfigApi.DeviceDockerConfig

type DeviceDockerConfigs struct {
	DeviceDockerConfigs []DeviceDockerConfig `json:"deviceDockerConfigs"`
}

func (d *Device) IsEmpty() bool {
	if cmp.Diff(Device{}, *d) == "" {
		return true
	}
	return false
}

func (d *Devices) IsEmpty() bool {
	if len(d.Devices) == 0 {
		return true
	}
	return false
}

func (d *Devices) GetIds() []string {
	out := make([]string, len(d.Devices))
	for _, dev := range d.Devices {
		out = append(out, dev.Metadata.ID)
	}
	return out
}

func (d *PatchDevice) IsEmpty() bool {
	if cmp.Diff(PatchDevice{}, *d) == "" {
		return true
	}
	return false
}

func (d *PatchDevice) IsValid() bool {
	if d.Spec.MACAddress == nil && d.Spec.MaintenanceWindow == nil {
		return false
	}

	return true
}

func (p *Pod) IsEmpty() bool {
	if cmp.Diff(Pod{}, *p) == "" {
		return true
	}
	return false
}

func (p *Pods) IsEmpty() bool {
	if len(p.Pods) == 0 {
		return true
	}
	return false
}

func (p *Pods) GetIds() []string {
	out := make([]string, len(p.Pods))
	for _, po := range p.Pods {
		out = append(out, po.Metadata.ID)
	}
	return out
}

func (d *Deployment) IsEmpty() bool {
	if cmp.Diff(Deployment{}, *d) == "" {
		return true
	}
	return false
}

func (d *Deployments) IsEmpty() bool {
	if len(d.Deployments) == 0 {
		return true
	}
	return false
}

func (d *Deployments) GetIds() []string {
	out := make([]string, len(d.Deployments))
	for _, dep := range d.Deployments {
		out = append(out, dep.Metadata.ID)
	}
	return out
}

func (d *CreateDeployment) IsEmpty() bool {
	if cmp.Diff(CreateDeployment{}, *d) == "" {
		return true
	}
	return false
}

func (d *PatchDeployment) IsEmpty() bool {
	if cmp.Diff(PatchDeployment{}, *d) == "" {
		return true
	}
	return false
}

func (a *CreateDeployment) IsValid() bool {
	if a.Spec.App == "" {
		return false
	}
	if a.Spec.Selector.MatchByLabels == nil {
		return false
	}
	if a.Spec.Template.Spec.Containers == nil {
		return false
	}

	return true
}

func (a *PatchDeployment) IsValid() bool {
	if a.Spec.App == "" {
		return false
	}
	if a.Spec.Selector.MatchByLabels == nil {
		return false
	}
	if a.Spec.Template.Spec.Containers == nil {
		return false
	}

	return true
}

func (a *Application) IsEmpty() bool {
	if cmp.Diff(Application{}, *a) == "" {
		return true
	}
	return false
}

func (a *Applications) IsEmpty() bool {
	if len(a.Applications) == 0 {
		return true
	}
	return false
}

func (a *CreateApplication) IsEmpty() bool {
	if cmp.Diff(CreateApplication{}, *a) == "" {
		return true
	}
	return false
}

func (a *CreateApplication) IsValid() bool {
	//Todo: Check if it's valid eg. mandatory fields
	return true
}

func (m *MaintenanceTask) IsEmpty() bool {
	if cmp.Diff(MaintenanceTask{}, *m) == "" {
		return true
	}
	return false
}

func (m *MaintenanceTasks) IsEmpty() bool {
	if len(m.MaintenanceTasks) == 0 {
		return true
	}
	return false
}

func (m *MaintenanceTasks) GetIds() []string {
	out := make([]string, len(m.MaintenanceTasks))
	for _, dev := range m.MaintenanceTasks {
		out = append(out, dev.Metadata.ID)
	}
	return out
}

func (d *DockerConfig) IsEmpty() bool {
	if cmp.Diff(DockerConfig{}, *d) == "" {
		return true
	}
	return false
}

func (d *DockerConfigs) IsEmpty() bool {
	if len(d.DockerConfigs) == 0 {
		return true
	}
	return false
}

func (d *DockerConfigs) GetIds() []string {
	out := make([]string, len(d.DockerConfigs))
	for _, con := range d.DockerConfigs {
		out = append(out, con.Metadata.ID)
	}
	return out
}

func (d *DeviceDockerConfig) IsEmpty() bool {
	if cmp.Diff(DeviceDockerConfig{}, *d) == "" {
		return true
	}
	return false
}

func (d *DeviceDockerConfigs) IsEmpty() bool {
	if len(d.DeviceDockerConfigs) == 0 {
		return true
	}
	return false
}

func (d *DeviceDockerConfigs) GetIds() []string {
	out := make([]string, len(d.DeviceDockerConfigs))
	for _, con := range d.DeviceDockerConfigs {
		out = append(out, con.Metadata.ID)
	}
	return out
}
