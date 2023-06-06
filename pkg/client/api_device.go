package client

import (
	"fmt"
	"strings"

	api "github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/errors"
	"github.com/grid-x/gxctl/pkg/filter"
)

type GetDevicesResponse struct {
	Device  api.Devices
	Profile string
	Err     error
}

func GetDeviceById(client *APIClient, id string, deviceIds []string) (api.Device, error) {
	if len(id) != 36 && deviceIds == nil {
		responseList, err := GetDevices(client, "", nil)
		if err != nil {
			return api.Device{}, err
		}
		for _, r := range responseList {
			deviceIds = append(deviceIds, r.Device.GetIds()...)
		}
	}

	deviceID, err := api.LookupID(id, deviceIds)
	if err != nil {
		return api.Device{}, err
	}

	endpoint := fmt.Sprintf("%s/%s", api.DevicesEndpoint, deviceID)
	response, err := client.GetRequest(endpoint)
	if err != nil {
		return api.Device{}, err
	}

	device, err := api.NewDevice(response, false)
	if err != nil {
		return device, err
	}

	return device, nil
}

func GetDevices(c *APIClient, sn string, f filter.Filter) (map[string]GetDevicesResponse, error) {
	var responses []RequestResult
	var err error

	if sn == "" {
		responses, err = c.GetMultiRequest(api.DevicesEndpoint)
	} else {
		responses, err = c.GetMultiRequest(fmt.Sprintf("%v?filter=serialnumber:%v", api.DevicesEndpoint, strings.ToUpper(sn)))
	}
	if err != nil {
		return nil, err
	}

	responseList := make(map[string]GetDevicesResponse)
	for _, response := range responses {
		resp := GetDevicesResponse{}
		resp.Profile = response.Profile
		if response.Err != nil {
			resp.Err = response.Err
			responseList[response.Profile] = resp
			continue
		}

		deviceList, err := api.NewDevices(response.Body, false)
		if err != nil {
			resp.Err = err
		}

		if f != nil {
			deviceList = filter.FilterDevices(deviceList, f)
		}

		if resp.Err == nil && deviceList.IsEmpty() {
			resp.Err = errors.E(
				errors.NotExists,
				"no devices found",
			)
		} else {
			resp.Device = deviceList
		}

		responseList[response.Profile] = resp
	}
	return responseList, nil
}
