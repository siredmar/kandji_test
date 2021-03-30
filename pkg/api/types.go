package api

import (
	"time"

	"github.com/google/go-cmp/cmp"
	types "github.com/grid-x/ds-api-types"
	applicationsApi "github.com/grid-x/ds-api-types/management/2018-11-27/application"
	deviceDockerConfigApi "github.com/grid-x/ds-api-types/management/2019-06-09/devicedockerconfigs"
	devicesApi "github.com/grid-x/ds-api-types/management/2019-06-13/device"
	podsApi "github.com/grid-x/ds-api-types/management/2019-08-17/pod"
	maintenanceApi "github.com/grid-x/ds-api-types/management/2019-11-04/maintenance"
	dockerConfigApi "github.com/grid-x/ds-api-types/management/2019-12-10/dockerconfigs"
	deploymentsApi "github.com/grid-x/ds-api-types/management/2020-08-29/deployments"
)

type Resource interface {
	Meta() *types.Metadata
}

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

type UpdateDevice devicesApi.UpdateRequest

type Pod podsApi.Pod

type Pods struct {
	Pods []Pod `json:"pods"`
}

type Deployment deploymentsApi.Deployment

type CreateDeployment deploymentsApi.CreateRequest

type UpdateDeployment deploymentsApi.UpdateRequest

type Deployments struct {
	Deployments []Deployment `json:"deployments"`
}

type Application applicationsApi.Application

type Applications struct {
	Applications []Application `json:"applications"`
}

type CreateApplication applicationsApi.CreateRequest

type MaintenanceTask maintenanceApi.MaintenanceTask

type MaintenanceTasks struct {
	MaintenanceTasks []MaintenanceTask `json:"maintenanceTasks"`
}

type CreateMaintenanceTask maintenanceApi.CreateRequest

func (cmt *CreateMaintenanceTask) Meta() *types.Metadata {
	return nil
}

type DockerConfig dockerConfigApi.DockerConfig

type DockerConfigs struct {
	DockerConfigs []DockerConfig `json:"dockerConfigs"`
}

type CreateDockerConfig dockerConfigApi.CreateRequest

type UpdateDockerConfig dockerConfigApi.UpdateRequest

type DeviceDockerConfig deviceDockerConfigApi.DeviceDockerConfig

type DeviceDockerConfigs struct {
	DeviceDockerConfigs []DeviceDockerConfig `json:"deviceDockerConfigs"`
}

func (d *Device) Meta() *types.Metadata {
	return &d.Metadata
}

func (d *Device) IsEmpty() bool {
	if cmp.Diff(Device{}, *d) == "" {
		return true
	}
	return false
}

func (d *Device) IsOnline() bool {
	return d.Status.LastHeartbeat != nil && time.Now().Sub(d.Status.LastHeartbeat.Time) < (2*time.Minute)
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

func (d *UpdateDevice) Meta() *types.Metadata {
	return nil
}

func (d *UpdateDevice) IsEmpty() bool {
	if cmp.Diff(UpdateDevice{}, *d) == "" {
		return true
	}
	return false
}

func (d *UpdateDevice) IsValid() bool {
	if d.Spec.MACAddress == nil && d.Spec.MaintenanceWindow == nil {
		return false
	}

	return true
}

func (p *Pod) Meta() *types.Metadata {
	return &p.Metadata
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

func (d *UpdateDeployment) Meta() *types.Metadata {
	return nil
}

func (d *Deployment) Meta() *types.Metadata {
	return &d.Metadata
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

func (d *UpdateDeployment) IsEmpty() bool {
	if cmp.Diff(UpdateDeployment{}, *d) == "" {
		return true
	}
	return false
}

func (a *CreateDeployment) IsValid() bool {
	if a.Spec.App == "" {
		return false
	}
	if a.Spec.Selector.MatchByLabels == nil && a.Spec.Selector.MatchByDeviceID == nil {
		return false
	}
	if a.Spec.Template.Spec.Containers == nil {
		return false
	}

	return true
}

func (a *UpdateDeployment) IsValid() bool {
	if a.Spec.App == "" {
		return false
	}
	if a.Spec.Selector.MatchByLabels == nil && a.Spec.Selector.MatchByDeviceID == nil {
		return false
	}
	if a.Spec.Template.Spec.Containers == nil {
		return false
	}

	return true
}

func (a *Application) Meta() *types.Metadata {
	return &a.Metadata
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

func (mt *MaintenanceTask) Meta() *types.Metadata {
	return &mt.Metadata
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

func (dc *UpdateDockerConfig) Meta() *types.Metadata {
	return nil
}

func (dc *DockerConfig) Meta() *types.Metadata {
	return &dc.Metadata
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

func (d *CreateDockerConfig) IsEmpty() bool {
	if cmp.Diff(CreateDockerConfig{}, *d) == "" {
		return true
	}
	return false
}

func (a *CreateDockerConfig) IsValid() bool {
	// Todo
	return true
}

func (d *UpdateDockerConfig) IsEmpty() bool {
	if cmp.Diff(UpdateDockerConfig{}, *d) == "" {
		return true
	}
	return false
}

func (a *UpdateDockerConfig) IsValid() bool {
	// Todo
	return true
}

func (ddc *DeviceDockerConfig) Meta() *types.Metadata {
	return &ddc.Metadata
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
