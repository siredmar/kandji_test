package api

import (
	"encoding/json"
)

func NewDevice(j []byte) (Device, error) {
	d := Device{}
	err := json.Unmarshal(j, &d)

	if err != nil {
		return d, err
	}
	return d, nil
}

func NewDevices(j []byte) (Devices, error) {
	d := Devices{}
	err := json.Unmarshal(j, &d)

	if err != nil {
		return d, err
	}
	return d, nil
}

func NewPod(j []byte) (Pod, error) {
	d := Pod{}
	err := json.Unmarshal(j, &d)

	if err != nil {
		return d, err
	}
	return d, nil
}

func NewPods(j []byte) (Pods, error) {
	d := Pods{}
	err := json.Unmarshal(j, &d)

	if err != nil {
		return d, err
	}
	return d, nil
}
