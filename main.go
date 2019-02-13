// Copyright © 2019 gridX <ops@gridx.de>
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

package main

import (
	"fmt"
	"os"

	cmd "github.com/grid-x/gxctl/cmd"
)

func main() {
	root := cmd.NewRoot()
	get := cmd.NewGet(root.Command)
	patch := cmd.NewPatch(root.Command)
	delete := cmd.NewDelete(root.Command)
	create := cmd.NewCreate(root.Command)
	cmd.NewCompletion(root.Command)
	cmd.NewGetDevices(get.Command)
	cmd.NewGetPods(get.Command)
	cmd.NewGetDeployments(get.Command)
	cmd.NewGetApplications(get.Command)
	cmd.NewGetMaintenance(get.Command)
	cmd.NewCreateApplication(create.Command)
	cmd.NewCreateDeployment(create.Command)
	cmd.NewPatchDevice(patch.Command)
	cmd.NewDeleteDeployment(delete.Command)
	cmd.NewDeleteApplication(delete.Command)

	if err := root.Command.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
