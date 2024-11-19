package action

import (
	"errors"
	"fmt"
	"time"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	"github.com/grid-x/gxctl/pkg/printer"
	"github.com/grid-x/gxctl/pkg/service"
)

var (
	ErrDeviceNotFound       = errors.New("device not found")
	ErrNotExactSerialNumber = errors.New("there are multiple devices with this serial number pattern. Please give an exact serial number")
	ErrWrongOwner           = errors.New("there are device logs settings for this device created by another user. If you want to override them, run the command again with the --change-owner flag")
)

func GetLogs(s *service.Service, id string, isSerialNumber bool, output string) error {
	printerCfg := printer.PrintConfig{
		OutputFormat: output,
	}

	if isSerialNumber {
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

	if err := s.Printer.Print(dl, printerCfg); err != nil {
		return fmt.Errorf("error printing device logs: %v", err)
	}

	return nil
}

func ListLogs(s *service.Service, output string) error {
	printerCfg := printer.PrintConfig{
		OutputFormat: output,
	}

	resp, err := s.Client.GetRequest(api.DeviceLogsListEndpoint)
	if err != nil {
		return fmt.Errorf("error getting device logs list: %v", err)
	}

	dl, err := api.NewDevicesLogs(resp, true)
	if err != nil {
		return fmt.Errorf("error creating device logs list: %v", err)
	}

	if err := s.Printer.Print(dl, printerCfg); err != nil {
		return fmt.Errorf("error printing the device logs list: %v", err)
	}

	return nil
}

func EnableLogs(s *service.Service, id string, isSerialNumber bool, logLevel, output string, duration time.Duration) error {
	printerCfg := printer.PrintConfig{
		OutputFormat: output,
	}

	if isSerialNumber {
		var err error
		id, err = getDeviceID(s, id)
		if err != nil {
			return fmt.Errorf("failed to get device ID: %v", err)
		}
	}

	var body api.DeviceLogs
	body.Spec.LogLevel = logLevel
	body.Spec.ExpiresAt = api.NewTime(time.Now().Add(duration))
	resp, err := s.Client.PostRequest(fmt.Sprintf("%s/%s", api.DeviceLogsEndpoint, id), &body)
	if err != nil {
		return fmt.Errorf("failed to enable device logs: %v", err)
	}

	deviceLogs, err := api.NewDeviceLogs(resp, true)
	if err != nil {
		return fmt.Errorf("error creating create device logs response: %v", err)
	}

	if err := s.Printer.Print(deviceLogs, printerCfg); err != nil {
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

func UpdateLogs(s *service.Service, id string, isSerialNumber, changeOwner bool, logLevel, output string, duration time.Duration) error {
	printerCfg := printer.PrintConfig{
		OutputFormat: output,
	}

	if isSerialNumber {
		var err error
		id, err = getDeviceID(s, id)
		if err != nil {
			return fmt.Errorf("error getting device ID: %v", err)
		}
	}

	var body api.DeviceLogs
	body.Spec.LogLevel = logLevel
	body.Spec.ExpiresAt = api.NewTime(time.Now().Add(duration))
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

	if err := s.Printer.Print(updatedDeviceLogs, printerCfg); err != nil {
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
