package api

import (
	"encoding/json"
)

func NewFullObjectMeta(j []byte) (FullObjectMeta, error) {
	d := FullObjectMeta{}
	err := json.Unmarshal(j, &d)

	if err != nil {
		return d, err
	}
	return d, nil
}

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

func NewPatchDevice(j []byte) (PatchDevice, error) {
	d := PatchDevice{}
	err := json.Unmarshal(j, &d)

	if err != nil {
		return d, err
	}
	return d, nil
}

func NewPod(j []byte) (Pod, error) {
	p := Pod{}
	err := json.Unmarshal(j, &p)

	if err != nil {
		return p, err
	}
	return p, nil
}

func NewPods(j []byte) (Pods, error) {
	p := Pods{}
	err := json.Unmarshal(j, &p)

	if err != nil {
		return p, err
	}
	return p, nil
}

func NewDeployment(j []byte) (Deployment, error) {
	d := Deployment{}
	err := json.Unmarshal(j, &d)

	if err != nil {
		return d, err
	}
	return d, nil
}

func NewDeployments(j []byte) (Deployments, error) {
	d := Deployments{}
	err := json.Unmarshal(j, &d)

	if err != nil {
		return d, err
	}
	return d, nil
}

func NewCreateDeployment(j []byte) (CreateDeployment, error) {
	d := CreateDeployment{}
	err := json.Unmarshal(j, &d)

	if err != nil {
		return d, err
	}
	return d, nil
}

func NewPatchDeployment(j []byte) (PatchDeployment, error) {
	d := PatchDeployment{}
	err := json.Unmarshal(j, &d)

	if err != nil {
		return d, err
	}
	return d, nil
}

func NewApplication(j []byte) (Application, error) {
	a := Application{}
	err := json.Unmarshal(j, &a)

	if err != nil {
		return a, err
	}
	return a, nil
}

func NewApplications(j []byte) (Applications, error) {
	a := Applications{}
	err := json.Unmarshal(j, &a)

	if err != nil {
		return a, err
	}
	return a, nil
}

func NewCreateApplication(j []byte) (CreateApplication, error) {
	a := CreateApplication{}
	err := json.Unmarshal(j, &a)

	if err != nil {
		return a, err
	}
	return a, nil
}
