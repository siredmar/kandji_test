package api

import (
	"time"

	"github.com/google/go-cmp/cmp"
	types "github.com/grid-x/ds-api-types"
	applicationsApi "github.com/grid-x/ds-api-types/management/2018-11-27/application"
	devicesApi "github.com/grid-x/ds-api-types/management/2019-06-13/device"
	podsApi "github.com/grid-x/ds-api-types/management/2019-08-17/pod"
	maintenanceApi "github.com/grid-x/ds-api-types/management/2019-11-04/maintenance"
	deploymentsApi "github.com/grid-x/ds-api-types/management/2020-08-29/deployments"
	deviceConfigMapApi "github.com/grid-x/ds-api-types/management/2021-03-10/deviceconfigmaps"
	devicelogsapi "github.com/grid-x/ds-api-types/management/2024-11-04/devicelogs"
)

type Resource interface {
	Meta() *types.Metadata
}

type FullObjectMeta struct {
	Meta struct {
		ID   string `json:"id,omitempty"`
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

type DeviceConfigMap deviceConfigMapApi.DeviceConfigMap
type DeviceConfigMaps struct {
	DeviceConfigMaps []DeviceConfigMap `json:"deviceConfigMaps"`
}
type UpdateDeviceConfigMap deviceConfigMapApi.UpdateRequest

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

func (d *Device) Meta() *types.Metadata {
	return &d.Metadata
}

func (d *Device) IsEmpty() bool {
	return cmp.Diff(Device{}, *d) == ""
}

func (d *Device) IsOnline() bool {
	return d.Status.LastHeartbeat != nil && time.Since(d.Status.LastHeartbeat.Time) < (2*time.Minute)
}

func (d *Devices) IsEmpty() bool {
	return len(d.Devices) == 0
}

func (d *Devices) GetIDs() []string {
	out := make([]string, 0, len(d.Devices))
	for _, dev := range d.Devices {
		out = append(out, dev.Metadata.ID)
	}
	return out
}

func (d *UpdateDevice) Meta() *types.Metadata {
	return nil
}

func (d *UpdateDevice) IsEmpty() bool {
	return cmp.Diff(UpdateDevice{}, *d) == ""
}

func (d *UpdateDevice) IsValid() bool {
	if d.Spec.MACAddress == nil && d.Spec.MaintenanceWindow == nil {
		return false
	}

	return true
}

func (dcm *DeviceConfigMap) Meta() *types.Metadata {
	return &dcm.Metadata
}

func (dcm *UpdateDeviceConfigMap) Meta() *types.Metadata {
	return nil
}

func (p *Pod) Meta() *types.Metadata {
	return &p.Metadata
}

func (p *Pod) IsEmpty() bool {
	return cmp.Diff(Pod{}, *p) == ""
}

func (p *Pods) IsEmpty() bool {
	return len(p.Pods) == 0
}

func (p *Pods) GetIDs() []string {
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
	return cmp.Diff(Deployment{}, *d) == ""
}

func (d *Deployments) IsEmpty() bool {
	return len(d.Deployments) == 0
}

func (d *Deployments) GetIDs() []string {
	out := make([]string, len(d.Deployments))
	for _, dep := range d.Deployments {
		out = append(out, dep.Metadata.ID)
	}
	return out
}

func (d *CreateDeployment) IsEmpty() bool {
	return cmp.Diff(CreateDeployment{}, *d) == ""
}

func (d *UpdateDeployment) IsEmpty() bool {
	return cmp.Diff(UpdateDeployment{}, *d) == ""
}

func (d *CreateDeployment) IsValid() bool {
	if d.Spec.App == "" {
		return false
	}
	if d.Spec.Selector.MatchByLabels == nil && d.Spec.Selector.MatchByDeviceID == nil {
		return false
	}
	if d.Spec.Template.Spec.Containers == nil {
		return false
	}

	return true
}

func (d *UpdateDeployment) IsValid() bool {
	if d.Spec.App == "" {
		return false
	}
	if d.Spec.Selector.MatchByLabels == nil && d.Spec.Selector.MatchByDeviceID == nil {
		return false
	}
	if d.Spec.Template.Spec.Containers == nil {
		return false
	}

	return true
}

func (a *Application) Meta() *types.Metadata {
	return &a.Metadata
}

func (a *Application) IsEmpty() bool {
	return cmp.Diff(Application{}, *a) == ""
}

func (a *Applications) IsEmpty() bool {
	return len(a.Applications) == 0
}

func (a *CreateApplication) IsEmpty() bool {
	return cmp.Diff(CreateApplication{}, *a) == ""
}

func (a *CreateApplication) IsValid() bool {
	// TODO: Check if it's valid eg. mandatory fields
	return true
}

func (m *MaintenanceTask) Meta() *types.Metadata {
	return &m.Metadata
}

func (m *MaintenanceTask) IsEmpty() bool {
	return cmp.Diff(MaintenanceTask{}, *m) == ""
}

func (m *MaintenanceTasks) IsEmpty() bool {
	return len(m.MaintenanceTasks) == 0
}

func (m *MaintenanceTasks) GetIDs() []string {
	out := make([]string, len(m.MaintenanceTasks))
	for _, dev := range m.MaintenanceTasks {
		out = append(out, dev.Metadata.ID)
	}
	return out
}

func NewTime(t time.Time) types.Time {
	return types.NewTime(t)
}

type DeviceLogs devicelogsapi.DeviceLogs
type DevicesLogs devicelogsapi.DeviceLogsList

func (r *DeviceLogs) Meta() *types.Metadata {
	return nil
}
