package api

import (
	"bytes"
	"encoding/json"
)

func NewFullObjectMeta(j []byte) (FullObjectMeta, error) {
	d := FullObjectMeta{}

	dec := json.NewDecoder(bytes.NewReader(j))
	err := dec.Decode(&d)
	return d, err
}

func NewDevice(j []byte, strict bool) (Device, error) {
	d := Device{}
	err := decode(j, &d, strict)
	return d, err
}

func NewDevices(j []byte, strict bool) (Devices, error) {
	d := Devices{}
	err := decode(j, &d, strict)
	return d, err
}

func NewUpdateDevice(j []byte, strict bool) (UpdateDevice, error) {
	d := UpdateDevice{}
	err := decode(j, &d, strict)
	return d, err
}

func NewDeviceConfigMap(j []byte, strict bool) (DeviceConfigMap, error) {
	dcm := DeviceConfigMap{}
	err := decode(j, &dcm, strict)
	return dcm, err
}

func NewDeviceConfigMaps(j []byte, strict bool) (DeviceConfigMaps, error) {
	dcm := DeviceConfigMaps{}
	err := decode(j, &dcm, strict)
	return dcm, err
}

func NewPod(j []byte, strict bool) (Pod, error) {
	p := Pod{}
	err := decode(j, &p, strict)
	return p, err
}

func NewPods(j []byte, strict bool) (Pods, error) {
	ps := Pods{}
	err := decode(j, &ps, strict)
	return ps, err
}

func NewDeployment(j []byte, strict bool) (Deployment, error) {
	d := Deployment{}
	err := decode(j, &d, strict)
	return d, err
}

func NewDeployments(j []byte, strict bool) (Deployments, error) {
	ds := Deployments{}
	err := decode(j, &ds, strict)
	return ds, err
}

func NewCreateDeployment(j []byte, strict bool) (CreateDeployment, error) {
	cd := CreateDeployment{}
	err := decode(j, &cd, strict)
	return cd, err
}

func NewUpdateDeployment(j []byte, strict bool) (UpdateDeployment, error) {
	ud := UpdateDeployment{}
	err := decode(j, &ud, strict)
	return ud, err
}

func NewApplication(j []byte, strict bool) (Application, error) {
	a := Application{}
	err := decode(j, &a, strict)
	return a, err
}

func NewApplications(j []byte, strict bool) (Applications, error) {
	as := Applications{}
	err := decode(j, &as, strict)
	return as, err
}

func NewCreateApplication(j []byte, strict bool) (CreateApplication, error) {
	ca := CreateApplication{}
	err := decode(j, &ca, strict)
	return ca, err
}

func NewMaintenanceTask(j []byte, strict bool) (MaintenanceTask, error) {
	mt := MaintenanceTask{}
	err := decode(j, &mt, strict)
	return mt, err
}

func NewMaintenanceTasks(j []byte, strict bool) (MaintenanceTasks, error) {
	mts := MaintenanceTasks{}
	err := decode(j, &mts, strict)
	return mts, err
}

func NewCreateMaintenanceTask(j []byte, strict bool) (CreateMaintenanceTask, error) {
	cmt := CreateMaintenanceTask{}
	err := decode(j, &cmt, strict)
	return cmt, err
}

func NewDevicesLogs(j []byte, strict bool) (DevicesLogs, error) {
	dl := DevicesLogs{}
	err := decode(j, &dl, strict)
	return dl, err
}

func NewDeviceLogs(j []byte, strict bool) (DeviceLogs, error) {
	dl := DeviceLogs{}
	err := decode(j, &dl, strict)
	return dl, err
}

func decode(j []byte, i any, strict bool) error {
	dec := json.NewDecoder(bytes.NewReader(j))
	if strict {
		dec.DisallowUnknownFields()
	}

	return dec.Decode(i)
}
