package api

import (
	"encoding/json"

	"github.com/ghodss/yaml"
)

func NewFullObjectMeta(j []byte) (FullObjectMeta, error) {
	d := FullObjectMeta{}

	b, err := yaml.YAMLToJSON(j)
	if err == nil {
		j = b
	}

	err = json.Unmarshal(j, &d)

	if err != nil {
		return d, err
	}

	return d, nil
}

func NewDevice(j []byte) (Device, error) {
	d := Device{}

	b, err := yaml.YAMLToJSON(j)
	if err == nil {
		j = b
	}

	err = json.Unmarshal(j, &d)

	if err != nil {
		return d, err
	}
	return d, nil
}

func NewDevices(j []byte) (Devices, error) {
	d := Devices{}

	b, err := yaml.YAMLToJSON(j)
	if err == nil {
		j = b
	}

	err = json.Unmarshal(j, &d)

	if err != nil {
		return d, err
	}
	return d, nil
}

func NewPatchDevice(j []byte) (PatchDevice, error) {
	d := PatchDevice{}

	b, err := yaml.YAMLToJSON(j)
	if err == nil {
		j = b
	}

	err = json.Unmarshal(j, &d)

	if err != nil {
		return d, err
	}

	return d, nil
}

func NewPod(j []byte) (Pod, error) {
	p := Pod{}

	b, err := yaml.YAMLToJSON(j)
	if err == nil {
		j = b
	}

	err = json.Unmarshal(j, &p)

	if err != nil {
		return p, err
	}
	return p, nil
}

func NewPods(j []byte) (Pods, error) {
	p := Pods{}

	b, err := yaml.YAMLToJSON(j)
	if err == nil {
		j = b
	}

	err = json.Unmarshal(j, &p)

	if err != nil {
		return p, err
	}
	return p, nil
}

func NewDeployment(j []byte) (Deployment, error) {
	d := Deployment{}

	b, err := yaml.YAMLToJSON(j)
	if err == nil {
		j = b
	}

	err = json.Unmarshal(j, &d)

	if err != nil {
		return d, err
	}
	return d, nil
}

func NewDeployments(j []byte) (Deployments, error) {
	d := Deployments{}

	b, err := yaml.YAMLToJSON(j)
	if err == nil {
		j = b
	}

	err = json.Unmarshal(j, &d)

	if err != nil {
		return d, err
	}
	return d, nil
}

func NewCreateDeployment(j []byte) (CreateDeployment, error) {
	d := CreateDeployment{}

	b, err := yaml.YAMLToJSON(j)
	if err == nil {
		j = b
	}

	err = json.Unmarshal(j, &d)

	if err != nil {
		return d, err
	}
	return d, nil
}

func NewPatchDeployment(j []byte) (PatchDeployment, error) {
	d := PatchDeployment{}

	b, err := yaml.YAMLToJSON(j)
	if err == nil {
		j = b
	}

	err = json.Unmarshal(j, &d)

	if err != nil {
		return d, err
	}
	return d, nil
}

func NewApplication(j []byte) (Application, error) {
	a := Application{}

	b, err := yaml.YAMLToJSON(j)
	if err == nil {
		j = b
	}

	err = json.Unmarshal(j, &a)

	if err != nil {
		return a, err
	}
	return a, nil
}

func NewApplications(j []byte) (Applications, error) {
	a := Applications{}

	b, err := yaml.YAMLToJSON(j)
	if err == nil {
		j = b
	}

	err = json.Unmarshal(j, &a)

	if err != nil {
		return a, err
	}
	return a, nil
}

func NewCreateApplication(j []byte) (CreateApplication, error) {
	a := CreateApplication{}

	b, err := yaml.YAMLToJSON(j)
	if err == nil {
		j = b
	}

	err = json.Unmarshal(j, &a)

	if err != nil {
		return a, err
	}
	return a, nil
}

func NewMaintenanceTask(j []byte) (MaintenanceTask, error) {
	m := MaintenanceTask{}
	err := json.Unmarshal(j, &m)

	if err != nil {
		return m, err
	}
	return m, nil
}

func NewMaintenanceTasks(j []byte) (MaintenanceTasks, error) {
	m := MaintenanceTasks{}
	err := json.Unmarshal(j, &m)

	if err != nil {
		return m, err
	}
	return m, nil
}

func NewCreateMaintenanceTask(j []byte) (CreateMaintenanceTask, error) {
	m := CreateMaintenanceTask{}
	err := json.Unmarshal(j, &m)

	if err != nil {
		return m, err
	}
	return m, nil
}
