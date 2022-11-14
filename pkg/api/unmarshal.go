package api

import (
	"bytes"
	"encoding/json"
)

func NewFullObjectMeta(j []byte) (FullObjectMeta, error) {
	d := FullObjectMeta{}

	dec := json.NewDecoder(bytes.NewReader(j))
	if err := dec.Decode(&d); err != nil {
		return d, err
	}

	return d, nil
}

func NewDevice(j []byte, strict bool) (Device, error) {
	d := Device{}

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

	dec := json.NewDecoder(bytes.NewReader(j))
	dec.DisallowUnknownFields() // Force errors
	if err := dec.Decode(&a); err != nil {
		return a, err
	}

	return a, nil
}

func NewMaintenanceTask(j []byte, strict bool) (MaintenanceTask, error) {
	m := MaintenanceTask{}

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

	dec := json.NewDecoder(bytes.NewReader(j))
	if strict {
		dec.DisallowUnknownFields()
	}
	if err := dec.Decode(&i); err != nil {
		return i, err
	}

	return i, nil
}
