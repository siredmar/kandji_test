package action

import (
	"errors"
	"fmt"
	"time"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	gerrors "github.com/grid-x/gxctl/pkg/errors"
	"github.com/grid-x/gxctl/pkg/printer"
	"github.com/grid-x/gxctl/pkg/service"
)

var (
	ErrDeviceNotFound       = errors.New("device not found")
	ErrNotExactSerialNumber = errors.New("there are multiple devices with this serial number pattern. Please give an exact serial number")
	ErrWrongOwner           = errors.New("there are device logs settings for this device created by another user. If you want to override them, run the command again with the --change-owner flag")
	ErrInvalidExpiry        = errors.New("Logs expiry must not exceed 1 month")
	ErrLogsAlreadyExists    = errors.New("Logs are already enabled for this device")

	maxExpiry = time.Hour * 24 * 30 // 1 month
)

func GetLogs(s *service.Service, id string, isSerialNumber bool, output string) error {
	printerCfg := printer.PrintConfig{
		OutputFormat: output,
	}

	var serialNumber string
	if isSerialNumber {
		serialNumber = id
		var err error
		id, err = getDeviceID(s, id)
		if err != nil {
			return fmt.Errorf("failed to get device ID: %v", err)
		}
	}

	dl, err := getDeviceLogsByDeviceID(s, id)
	if err != nil {
		return fmt.Errorf("error getting device logs by device ID: %v", err)
	}

	if serialNumber == "" {
		device, err := client.GetDeviceByID(s.Client, dl.Metadata.ID, nil)
		if err != nil {
			return fmt.Errorf("error getting a device: %v", err)
		}
		serialNumber = device.Spec.Serialnumber
	}

	if err := s.Printer.Print(api.NewDeviceLogsWithSerialNumber(dl, serialNumber), printerCfg); err != nil {
		return fmt.Errorf("error printing device logs: %v", err)
	}

	return nil
}

func ListLogs(s *service.Service, output string) error {
	printerCfg := printer.PrintConfig{
		OutputFormat: output,
	}

	resp, err := s.Client.GetRequest(api.DeviceLogsEndpoint)
	if err != nil {
		return fmt.Errorf("error getting device logs list: %v", err)
	}

	dl, err := api.NewDevicesLogs(resp, true)
	if err != nil {
		return fmt.Errorf("error creating device logs list: %v", err)
	}

	devicesLogsWithSerialNumber := make(api.DevicesLogsWithSerialNumber, len(dl.Items))
	for i, d := range dl.Items {
		device, err := client.GetDeviceByID(s.Client, d.Metadata.ID, nil)
		if err != nil {
			return fmt.Errorf("error getting a device: %v", err)
		}
		devicesLogsWithSerialNumber[i] = api.NewDeviceLogsWithSerialNumber(api.DeviceLogs(d), device.Spec.Serialnumber)
	}

	if err := s.Printer.Print(devicesLogsWithSerialNumber, printerCfg); err != nil {
		return fmt.Errorf("error printing the device logs list: %v", err)
	}

	return nil
}

func EnableLogs(s *service.Service, id string, isSerialNumber bool, logLevel, output string, expiry time.Duration) error {
	if expiry > maxExpiry {
		return ErrInvalidExpiry
	}

	printerCfg := printer.PrintConfig{
		OutputFormat: output,
	}

	var serialNumber string
	if isSerialNumber {
		serialNumber = id
		var err error
		id, err = getDeviceID(s, id)
		if err != nil {
			return fmt.Errorf("failed to get device ID: %v", err)
		}
	}

	var body api.DeviceLogs
	body.Spec.LogLevel = logLevel
	body.Spec.ExpiresAt = api.NewTime(time.Now().Add(expiry))
	resp, err := s.Client.PostRequest(fmt.Sprintf("%s/%s", api.DeviceLogsEndpoint, id), &body)
	if err != nil {
		var gerr *gerrors.Error
		if errors.As(err, &gerr) && gerr.Kind == gerrors.Exist {
			return ErrLogsAlreadyExists
		}
		return fmt.Errorf("failed to enable device logs: %v", err)
	}

	deviceLogs, err := api.NewDeviceLogs(resp, true)
	if err != nil {
		return fmt.Errorf("error creating create device logs response: %v", err)
	}

	if serialNumber == "" {
		device, err := client.GetDeviceByID(s.Client, deviceLogs.Metadata.ID, nil)
		if err != nil {
			return fmt.Errorf("error getting a device: %v", err)
		}
		serialNumber = device.Spec.Serialnumber
	}

	if err := s.Printer.Print(api.NewDeviceLogsWithSerialNumber(deviceLogs, serialNumber), printerCfg); err != nil {
		return fmt.Errorf("error printing device logs: %v", err)
	}
	return nil
}

func DisableLogs(s *service.Service, id string, isSerialNumber bool) error {
	if isSerialNumber {
		var err error
		id, err = getDeviceID(s, id)
		if err != nil {
			return fmt.Errorf("failed to get device ID: %v", err)
		}
	}

	if _, err := s.Client.DeleteRequest(api.DeviceLogsEndpoint, id); err != nil {
		return fmt.Errorf("failed to delete device logs: %v", err)
	}

	fmt.Printf("device logs settings for device %s were removed\n", id)
	return nil
}

func UpdateLogs(s *service.Service, id string, isSerialNumber, changeOwner bool, logLevel, output string, expiry time.Duration) error {
	if expiry > maxExpiry {
		return ErrInvalidExpiry
	}

	printerCfg := printer.PrintConfig{
		OutputFormat: output,
	}

	var serialNumber string
	if isSerialNumber {
		serialNumber = id
		var err error
		id, err = getDeviceID(s, id)
		if err != nil {
			return fmt.Errorf("error getting device ID: %v", err)
		}
	}

	var body api.DeviceLogs
	body.Spec.LogLevel = logLevel
	body.Spec.ExpiresAt = api.NewTime(time.Now().Add(expiry))
	if !changeOwner {
		dl, err := getDeviceLogsByDeviceID(s, id)
		if err != nil {
			return fmt.Errorf("error getting device logs: %v", err)
		}

		token, err := s.Client.GetToken()
		if err != nil {
			return fmt.Errorf("error getting user token: %v", err)
		}

		if oldOwner := dl.Spec.Owner; oldOwner != token.Email() {
			return ErrWrongOwner
		}
	}

	resp, err := s.Client.PatchRequest(api.DeviceLogsEndpoint, &body, id)
	if err != nil {
		return fmt.Errorf("error updating device logs: %v", err)
	}

	updatedDeviceLogs, err := api.NewDeviceLogs(resp, true)
	if err != nil {
		return fmt.Errorf("error creating update device logs response: %v", err)
	}

	if serialNumber == "" {
		device, err := client.GetDeviceByID(s.Client, updatedDeviceLogs.Metadata.ID, nil)
		if err != nil {
			return fmt.Errorf("error getting a device: %v", err)
		}
		serialNumber = device.Spec.Serialnumber
	}

	if err := s.Printer.Print(api.NewDeviceLogsWithSerialNumber(updatedDeviceLogs, serialNumber), printerCfg); err != nil {
		return fmt.Errorf("error printing updated device logs: %v", err)
	}

	return nil
}

func getDeviceLogsByDeviceID(s *service.Service, deviceID string) (api.DeviceLogs, error) {
	resp, err := s.Client.GetRequest(fmt.Sprintf("%s/%s", api.DeviceLogsEndpoint, deviceID))
	if err != nil {
		return api.DeviceLogs{}, fmt.Errorf("error getting device logs: %v", err)
	}

	dl, err := api.NewDeviceLogs(resp, true)
	if err != nil {
		return api.DeviceLogs{}, fmt.Errorf("error creating device logs: %v", err)
	}

	return dl, nil
}

func getDeviceID(s *service.Service, serialNumber string) (string, error) {
	profileToDevices, err := client.GetDevices(s.Client, serialNumber, nil)
	if err != nil {
		return "", fmt.Errorf("error getting devices by serial number: %v", err)
	}

	if len(profileToDevices) != 1 {
		return "", ErrNotExactSerialNumber
	}
	v, ok := profileToDevices[s.Config.Auth.CurrentProfile]
	if !ok {
		return "", ErrDeviceNotFound
	}

	if v.Err != nil {
		return "", fmt.Errorf("error getting device: %v", err)
	}

	if len(v.Device.Devices) != 1 {
		return "", ErrNotExactSerialNumber
	}

	return v.Device.Devices[0].Metadata.ID, nil

}
