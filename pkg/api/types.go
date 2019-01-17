package api

import (
	devicesApi "github.com/grid-x/ds-api/api/management/2018-11-02/devices"
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
