package api

import (
	"github.com/google/go-cmp/cmp"

	devicesApi "github.com/grid-x/ds-api/api/management/2018-11-02/devices"
	applicationsApi "github.com/grid-x/ds-api/api/management/2018-11-27/application"
	deploymentsApi "github.com/grid-x/ds-api/api/management/2018-11-28/deployments"
	podsApi "github.com/grid-x/ds-api/api/management/2018-12-04/pods"
)

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
