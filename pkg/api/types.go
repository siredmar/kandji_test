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

type Pod podsApi.Pod

type Pods struct {
	Pods []Pod `json:"pods"`
}

type Deployment deploymentsApi.Deployment

type CreateDeployment deploymentsApi.CreateRequest

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

func (a *CreateDeployment) IsValid() bool {
	//Todo: Check if it's valid eg. mandatory fields
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
