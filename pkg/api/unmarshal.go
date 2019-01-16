package api

import (
	"encoding/json"
	"fmt"
	devicesApi "github.com/grid-x/ds-api/api/management/2018-11-02/devices"
	podsApi "github.com/grid-x/ds-api/api/management/2018-12-04/pods"
	"os"
)

type Device devicesApi.Device

type Devices struct {
	Devices []Device `json:"devices"`
}

type Pod podsApi.Pod

type Pods struct {
	Pods []Pod `json:"pods"`
}

func (d Device) InitFromJSON(j []byte) Device {
	err := json.Unmarshal(j, &d)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return d
}

func (d Devices) InitFromJSON(j []byte) Devices {
	err := json.Unmarshal(j, &d)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return d
}

func (p Pod) InitFromJSON(j []byte) Pod {
	err := json.Unmarshal(j, &p)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return p
}

func (p Pods) InitFromJSON(j []byte) Pods {
	err := json.Unmarshal(j, &p)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return p
}
