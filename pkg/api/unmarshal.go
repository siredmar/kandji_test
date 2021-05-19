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
	if err := dec.Decode(&d); err != nil {
		return d, err
	}

	return d, nil
}

func NewDevice(j []byte, strict bool) (Device, error) {
	d := Device{}

	b, err := yaml.YAMLToJSON(j)
	if err == nil {
		j = b
	}

	dec := json.NewDecoder(bytes.NewReader(j))
	if strict {
		dec.DisallowUnknownFields()
	}
	if err := dec.Decode(&d); err != nil {
		return d, err
	}

	return d, nil
}

func NewDevices(j []byte, strict bool) (Devices, error) {
	d := Devices{}

	b, err := yaml.YAMLToJSON(j)
	if err == nil {
		j = b
	}

	dec := json.NewDecoder(bytes.NewReader(j))
	if strict {
		dec.DisallowUnknownFields()
	}
	if err := dec.Decode(&d); err != nil {
		return d, err
	}

	return d, nil
}

func NewUpdateDevice(j []byte, strict bool) (UpdateDevice, error) {
	d := UpdateDevice{}

	b, err := yaml.YAMLToJSON(j)
	if err == nil {
		j = b
	}

	dec := json.NewDecoder(bytes.NewReader(j))
	if strict {
		dec.DisallowUnknownFields()
	}
	if err := dec.Decode(&d); err != nil {
		return d, err
	}

	return d, nil
}

func NewDeviceConfigMap(j []byte, strict bool) (DeviceConfigMap, error) {
	dcm := DeviceConfigMap{}

	b, err := yaml.YAMLToJSON(j)
	if err == nil {
		j = b
	}

	dec := json.NewDecoder(bytes.NewReader(j))
	if strict {
		dec.DisallowUnknownFields()
	}
	if err := dec.Decode(&dcm); err != nil {
		return dcm, err
	}

	return dcm, nil
}

func NewDeviceConfigMaps(j []byte, strict bool) (DeviceConfigMaps, error) {
	dcm := DeviceConfigMaps{}

	b, err := yaml.YAMLToJSON(j)
	if err == nil {
		j = b
	}

	dec := json.NewDecoder(bytes.NewReader(j))
	if strict {
		dec.DisallowUnknownFields()
	}
	if err := dec.Decode(&dcm); err != nil {
		return dcm, err
	}

	return dcm, nil
}

func NewPod(j []byte, strict bool) (Pod, error) {
	p := Pod{}

	b, err := yaml.YAMLToJSON(j)
	if err == nil {
		j = b
	}

	dec := json.NewDecoder(bytes.NewReader(j))
	if strict {
		dec.DisallowUnknownFields()
	}
	if err := dec.Decode(&p); err != nil {
		return p, err
	}

	return p, nil
}

func NewPods(j []byte, strict bool) (Pods, error) {
	p := Pods{}

	b, err := yaml.YAMLToJSON(j)
	if err == nil {
		j = b
	}

	dec := json.NewDecoder(bytes.NewReader(j))
	if strict {
		dec.DisallowUnknownFields()
	}
	if err := dec.Decode(&p); err != nil {
		return p, err
	}

	return p, nil
}

func NewDeployment(j []byte, strict bool) (Deployment, error) {
	d := Deployment{}

	b, err := yaml.YAMLToJSON(j)
	if err == nil {
		j = b
	}

	dec := json.NewDecoder(bytes.NewReader(j))
	if strict {
		dec.DisallowUnknownFields()
	}
	if err := dec.Decode(&d); err != nil {
		return d, err
	}

	return d, nil
}

func NewDeployments(j []byte, strict bool) (Deployments, error) {
	d := Deployments{}

	b, err := yaml.YAMLToJSON(j)
	if err == nil {
		j = b
	}

	dec := json.NewDecoder(bytes.NewReader(j))
	if strict {
		dec.DisallowUnknownFields()
	}
	if err := dec.Decode(&d); err != nil {
		return d, err
	}

	return d, nil
}

func NewCreateDeployment(j []byte, strict bool) (CreateDeployment, error) {
	d := CreateDeployment{}

	b, err := yaml.YAMLToJSON(j)
	if err == nil {
		j = b
	}

	dec := json.NewDecoder(bytes.NewReader(j))
	if strict {
		dec.DisallowUnknownFields()
	}
	if err := dec.Decode(&d); err != nil {
		return d, err
	}

	return d, nil
}

func NewUpdateDeployment(j []byte, strict bool) (UpdateDeployment, error) {
	d := UpdateDeployment{}

	b, err := yaml.YAMLToJSON(j)
	if err == nil {
		j = b
	}

	dec := json.NewDecoder(bytes.NewReader(j))
	if strict {
		dec.DisallowUnknownFields()
	}
	if err := dec.Decode(&d); err != nil {
		return d, err
	}

	return d, nil
}

func NewApplication(j []byte, strict bool) (Application, error) {
	a := Application{}

	b, err := yaml.YAMLToJSON(j)
	if err == nil {
		j = b
	}

	dec := json.NewDecoder(bytes.NewReader(j))
	if strict {
		dec.DisallowUnknownFields()
	}
	if err := dec.Decode(&a); err != nil {
		return a, err
	}

	return a, nil
}

func NewApplications(j []byte, strict bool) (Applications, error) {
	a := Applications{}

	b, err := yaml.YAMLToJSON(j)
	if err == nil {
		j = b
	}

	dec := json.NewDecoder(bytes.NewReader(j))
	if strict {
		dec.DisallowUnknownFields()
	}
	if err := dec.Decode(&a); err != nil {
		return a, err
	}

	return a, nil
}

func NewCreateApplication(j []byte, strict bool) (CreateApplication, error) {
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

func NewMaintenanceTask(j []byte, strict bool) (MaintenanceTask, error) {
	m := MaintenanceTask{}

	b, err := yaml.YAMLToJSON(j)
	if err == nil {
		j = b
	}

	dec := json.NewDecoder(bytes.NewReader(j))
	if strict {
		dec.DisallowUnknownFields()
	}
	if err := dec.Decode(&m); err != nil {
		return m, err
	}

	return m, nil
}

func NewMaintenanceTasks(j []byte, strict bool) (MaintenanceTasks, error) {
	m := MaintenanceTasks{}

	b, err := yaml.YAMLToJSON(j)
	if err == nil {
		j = b
	}

	dec := json.NewDecoder(bytes.NewReader(j))
	if strict {
		dec.DisallowUnknownFields()
	}
	if err := dec.Decode(&m); err != nil {
		return m, err
	}

	return m, nil
}

func NewCreateMaintenanceTask(j []byte, strict bool) (CreateMaintenanceTask, error) {
	m := CreateMaintenanceTask{}

	b, err := yaml.YAMLToJSON(j)
	if err == nil {
		j = b
	}

	dec := json.NewDecoder(bytes.NewReader(j))
	if strict {
		dec.DisallowUnknownFields()
	}
	if err := dec.Decode(&m); err != nil {
		return m, err
	}

	return m, nil
}

func decode(j []byte, i interface{}, strict bool) (interface{}, error) {
	b, err := yaml.YAMLToJSON(j)
	if err == nil {
		j = b
	}

	dec := json.NewDecoder(bytes.NewReader(j))
	if strict {
		dec.DisallowUnknownFields()
	}
	if err := dec.Decode(&i); err != nil {
		return i, err
	}

	return i, nil
}
