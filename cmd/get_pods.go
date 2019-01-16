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
)

// podCmd represents the pod command
var podsCmd = &cobra.Command{
	Use:     "pods",
	Aliases: []string{"pod"},
	Short:   "get pods",
	Long:    `Prints a list of all pods you have access to`,
	Run: func(cmd *cobra.Command, args []string) {
		if GetCmdDeviceID != "" {
			//Get pods by Device ID
			pods, err := getPodByDeviceId(GetCmdDeviceID)

			if err == nil {
				printPods(pods)
			}
		} else if len(args) > 0 {
			//Get pods
			for _, a := range args {
				pod, err := getPodById(a)

				if err == nil {
					printPod(pod)
				}
			}
		} else {
			//List pods
			pods, err := getPods()

			if err == nil {
				printPods(pods)
			}
		}
	},
}

func getPods() (api.Pods, error) {
	response := client.Request(api.PodsEndpoint)
	pods := api.Pods{}.InitFromJSON(response)

	if len(pods.Pods) == 0 {
		return pods, api.NotFoundError(api.ErrorDetails{Command: "pods"})
	}

	return pods, nil
}

func getPodById(id string) (api.Pod, error) {
	response := client.Request(api.PodsEndpoint + "/" + id)
	pod := api.Pod{}.InitFromJSON(response)

	if pod.UUID != id {
		return pod, api.NotFoundError(api.ErrorDetails{Command: "pod", Id: id})
	}

	return pod, nil
}

func getPodByDeviceId(id string) (api.Pods, error) {
	return api.Pods{}, api.NotImplementedError()
	/*r
	esponse := apiClient.Request(api.DevicesEndpoint + "/" + id + "/pods")
	pods := api.Pods{}.InitFromJSON(response)

	if len(pods.Pods) == 0 {
		api.NotFoundError(api.ErrorDetails{Command: "pods", Id: id})
		return pods, errors.New("Not Found")
	}

	return pods, nil
	*/
}

func printPod(p api.Pod) {
	consolePrinter.PodConsoleOutput{}.Map(p).Print()
}

func printPods(p api.Pods) {
	consolePrinter.PodsConsoleOutput{}.Map(p).Sort().Print()
}

func init() {
	getCmd.AddCommand(podsCmd)
}
