// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"github.com/spf13/cobra"

	api "github.com/grid-x/gxctl/pkg/api"
	client "github.com/grid-x/gxctl/pkg/client"
	consolePrinter "github.com/grid-x/gxctl/pkg/printer/console"
	jsonPrinter "github.com/grid-x/gxctl/pkg/printer/json"
)

// devicesCmd represents the devices command
var devicesCmd = &cobra.Command{
	Use:     "devices",
	Aliases: []string{"device"},
	Short:   "get devices",
	Long:    `Prints a list of all devices you have access to`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			//Get devices
			for _, a := range args {
				device, err := getDeviceById(a)

				if err == nil {
					printDevice(device, GetCmdOutputType)
				}
			}
		} else if GetCmdDeviceID != "" {
			//Get device
			device, err := getDeviceById(GetCmdDeviceID)

			if err == nil {
				printDevice(device, GetCmdOutputType)
			}
		} else {
			//List devices
			devices, err := getDevices()

			if err == nil {
				printDevices(devices, GetCmdOutputType)
			}
		}
	},
}

func getDevices() (api.Devices, error) {
	response := client.Request(api.DevicesEndpoint)
	devices := api.Devices{}.InitFromJSON(response)

	if len(devices.Devices) == 0 {
		return devices, api.NotFoundError(api.ErrorDetails{Command: "devices"})
	}

	return devices, nil
}

func getDeviceById(id string) (api.Device, error) {
	response := client.Request(api.DevicesEndpoint + "/" + id)
	device := api.Device{}.InitFromJSON(response)

	if device.ID != id {
		return device, api.NotFoundError(api.ErrorDetails{Command: "device", Id: id})
	}

	return device, nil
}

func printDevice(d api.Device, outputFormat string) {
	if outputFormat == "json" {
		jsonPrinter.JSONOutput(d)
	} else {
		consolePrinter.DeviceConsoleOutput{}.Map(d).Print()
	}
}

func printDevices(d api.Devices, outputFormat string) {
	if outputFormat == "json" {
		jsonPrinter.JSONOutput(d)
	} else {
		consolePrinter.DevicesConsoleOutput{}.Map(d).Sort().Print()
	}
}

func init() {
	getCmd.AddCommand(devicesCmd)
}
