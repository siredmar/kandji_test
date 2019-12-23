package api

import (
	"bytes"
	"encoding/json"

	"github.com/ghodss/yaml"
)

func NewFullObjectMeta(j []byte) (FullObjectMeta, error) {
	d := FullObjectMeta{}

	b, err := yaml.YAMLToJSON(j)
	if err == nil {
		j = b
	}

	dec := json.NewDecoder(bytes.NewReader(j))
	// allow unknown fields
	if err := dec.Decode(&d); err != nil {
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

	dec := json.NewDecoder(bytes.NewReader(j))
	dec.DisallowUnknownFields() // Force errors
	if err := dec.Decode(&d); err != nil {
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

	dec := json.NewDecoder(bytes.NewReader(j))
	dec.DisallowUnknownFields() // Force errors
	if err := dec.Decode(&d); err != nil {
		return d, err
	}

	return d, nil
}

func NewUpdateDevice(j []byte) (UpdateDevice, error) {
	d := UpdateDevice{}

	b, err := yaml.YAMLToJSON(j)
	if err == nil {
		j = b
	}

	dec := json.NewDecoder(bytes.NewReader(j))
	dec.DisallowUnknownFields() // Force errors
	if err := dec.Decode(&d); err != nil {
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

	dec := json.NewDecoder(bytes.NewReader(j))
	dec.DisallowUnknownFields() // Force errors
	if err := dec.Decode(&p); err != nil {
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

	dec := json.NewDecoder(bytes.NewReader(j))
	dec.DisallowUnknownFields() // Force errors
	if err := dec.Decode(&p); err != nil {
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

	dec := json.NewDecoder(bytes.NewReader(j))
	dec.DisallowUnknownFields() // Force errors
	if err := dec.Decode(&d); err != nil {
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

	dec := json.NewDecoder(bytes.NewReader(j))
	dec.DisallowUnknownFields() // Force errors
	if err := dec.Decode(&d); err != nil {
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

	dec := json.NewDecoder(bytes.NewReader(j))
	dec.DisallowUnknownFields() // Force errors
	if err := dec.Decode(&d); err != nil {
		return d, err
	}

	return d, nil
}

func NewUpdateDeployment(j []byte) (UpdateDeployment, error) {
	d := UpdateDeployment{}

	b, err := yaml.YAMLToJSON(j)
	if err == nil {
		j = b
	}

	dec := json.NewDecoder(bytes.NewReader(j))
	dec.DisallowUnknownFields() // Force errors
	if err := dec.Decode(&d); err != nil {
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

	dec := json.NewDecoder(bytes.NewReader(j))
	dec.DisallowUnknownFields() // Force errors
	if err := dec.Decode(&a); err != nil {
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

	dec := json.NewDecoder(bytes.NewReader(j))
	dec.DisallowUnknownFields() // Force errors
	if err := dec.Decode(&a); err != nil {
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

	dec := json.NewDecoder(bytes.NewReader(j))
	dec.DisallowUnknownFields() // Force errors
	if err := dec.Decode(&a); err != nil {
		return a, err
	}

	return a, nil
}

func NewMaintenanceTask(j []byte) (MaintenanceTask, error) {
	m := MaintenanceTask{}

	b, err := yaml.YAMLToJSON(j)
	if err == nil {
		j = b
	}

	dec := json.NewDecoder(bytes.NewReader(j))
	dec.DisallowUnknownFields() // Force errors
	if err := dec.Decode(&m); err != nil {
		return m, err
	}

	return m, nil
}

func NewMaintenanceTasks(j []byte) (MaintenanceTasks, error) {
	m := MaintenanceTasks{}

	b, err := yaml.YAMLToJSON(j)
	if err == nil {
		j = b
	}

	dec := json.NewDecoder(bytes.NewReader(j))
	dec.DisallowUnknownFields() // Force errors
	if err := dec.Decode(&m); err != nil {
		return m, err
	}

	return m, nil
}

func NewCreateMaintenanceTask(j []byte) (CreateMaintenanceTask, error) {
	m := CreateMaintenanceTask{}

	b, err := yaml.YAMLToJSON(j)
	if err == nil {
		j = b
	}

	dec := json.NewDecoder(bytes.NewReader(j))
	dec.DisallowUnknownFields() // Force errors
	if err := dec.Decode(&m); err != nil {
		return m, err
	}

	return m, nil
}

func NewDockerConfig(j []byte) (DockerConfig, error) {
	d := DockerConfig{}

	b, err := yaml.YAMLToJSON(j)
	if err == nil {
		j = b
	}

	dec := json.NewDecoder(bytes.NewReader(j))
	dec.DisallowUnknownFields() // Force errors
	if err := dec.Decode(&d); err != nil {
		return d, err
	}

	return d, nil
}

func NewCreateDockerConfig(j []byte) (CreateDockerConfig, error) {
	d := CreateDockerConfig{}

	b, err := yaml.YAMLToJSON(j)
	if err == nil {
		j = b
	}

	dec := json.NewDecoder(bytes.NewReader(j))
	dec.DisallowUnknownFields() // Force errors
	if err := dec.Decode(&d); err != nil {
		return d, err
	}

	return d, nil
}

func NewDockerConfigs(j []byte) (DockerConfigs, error) {
	d := DockerConfigs{}

	b, err := yaml.YAMLToJSON(j)
	if err == nil {
		j = b
	}

	dec := json.NewDecoder(bytes.NewReader(j))
	dec.DisallowUnknownFields() // Force errors
	if err := dec.Decode(&d); err != nil {
		return d, err
	}

	return d, nil
}

func NewDeviceDockerConfig(j []byte) (DeviceDockerConfig, error) {
	d := DeviceDockerConfig{}

	b, err := yaml.YAMLToJSON(j)
	if err == nil {
		j = b
	}

	dec := json.NewDecoder(bytes.NewReader(j))
	dec.DisallowUnknownFields() // Force errors
	if err := dec.Decode(&d); err != nil {
		return d, err
	}

	return d, nil
}

func NewDeviceDockerConfigs(j []byte) (DeviceDockerConfigs, error) {
	d := DeviceDockerConfigs{}

	b, err := yaml.YAMLToJSON(j)
	if err == nil {
		j = b
	}

	dec := json.NewDecoder(bytes.NewReader(j))
	dec.DisallowUnknownFields() // Force errors
	if err := dec.Decode(&d); err != nil {
		return d, err
	}

	return d, nil
}

func NewCleanupConfig(j []byte) (CleanupConfig, error) {
	c := CleanupConfig{}

	b, err := yaml.YAMLToJSON(j)
	if err == nil {
		j = b
	}

	dec := json.NewDecoder(bytes.NewReader(j))
	dec.DisallowUnknownFields() // Force errors
	if err := dec.Decode(&c); err != nil {
		return c, err
	}

	return c, nil
}

func NewCreateCleanupConfig(j []byte) (CreateCleanupConfig, error) {
	d := CreateCleanupConfig{}

	b, err := yaml.YAMLToJSON(j)
	if err == nil {
		j = b
	}

	dec := json.NewDecoder(bytes.NewReader(j))
	dec.DisallowUnknownFields() // Force errors
	if err := dec.Decode(&d); err != nil {
		return d, err
	}

	return d, nil
}

func NewCleanupConfigs(j []byte) (CleanupConfigs, error) {
	c := CleanupConfigs{}

	b, err := yaml.YAMLToJSON(j)
	if err == nil {
		j = b
	}

	dec := json.NewDecoder(bytes.NewReader(j))
	dec.DisallowUnknownFields() // Force errors
	if err := dec.Decode(&c); err != nil {
		return c, err
	}

	return c, nil
}

func decode(j []byte, i interface{}) (interface{}, error) {
	b, err := yaml.YAMLToJSON(j)
	if err == nil {
		j = b
	}

	dec := json.NewDecoder(bytes.NewReader(j))
	dec.DisallowUnknownFields() // Force errors
	if err := dec.Decode(&i); err != nil {
		return i, err
	}

	return i, nil
}
