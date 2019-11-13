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

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/grid-x/gxctl/cmd"
	"github.com/grid-x/gxctl/pkg/client"
	"github.com/grid-x/gxctl/pkg/printer"
)

var (
	cfgFile string
	profile string
	staging bool
	conf    client.AuthConfig
)

func main() {
	root := cmd.NewRoot()

	client := client.NewAPIClient(&staging, &conf, &profile)
	printer := printer.NewPrinter()

	// Get
	get := cmd.NewGet(root.Command)
	cmd.NewGetDevices(get.Command, client, printer)
	cmd.NewGetPods(get.Command, client, printer)
	cmd.NewGetDeployments(get.Command, client, printer)
	cmd.NewGetApplications(get.Command, client, printer)

	// Lablel
	label := cmd.NewLabel(root.Command)
	cmd.NewLabelDevice(label.Command, client)

	// Delete
	delete := cmd.NewDelete(root.Command)
	cmd.NewDeleteDeployment(delete.Command, client)
	cmd.NewDeleteApplication(delete.Command, client)

	// Patch
	patch := cmd.NewPatch(root.Command, client)
	cmd.NewPatchDevice(patch.Command, client)
	cmd.NewPatchDeployment(patch.Command, client)

	// Create
	create := cmd.NewCreate(root.Command, client)
	cmd.NewCreateApplication(create.Command, client)
	cmd.NewCreateDeployment(create.Command, client)

	// Config
	config := cmd.NewConfig(root.Command)
	configGet := cmd.NewConfigGet(config.Command)
	configCreate := cmd.NewConfigCreate(config.Command)
	configDelete := cmd.NewConfigDelete(config.Command)
	cmd.NewConfigGetDocker(configGet.Command, client, printer)
	cmd.NewConfigCreateDockerAWS(configCreate.Command, client)
	cmd.NewConfigDeleteDocker(configDelete.Command, client)

	// Others
	cmd.NewCopy(root.Command, client)
	cmd.NewSSH(root.Command, client)
	cmd.NewSyslog(root.Command, client)
	cmd.NewPortForward(root.Command, client)
	cmd.NewCompletion(root.Command)
	cmd.NewRestart(root.Command, client)

	// Init config
	root.Command.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	root.Command.PersistentFlags().StringVar(&profile, "profile", "", "profile to use")
	root.Command.PersistentFlags().BoolVar(&staging, "staging", false, "should point to staging api")
	cobra.OnInitialize(initConfig)

	if err := root.Command.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
		viper.AddConfigPath("$HOME/.gxctl")
		viper.AddConfigPath(".") // optionally look for config in the working directory
	}

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println(err)
	}

	if err := viper.Unmarshal(&conf); err != nil {
		fmt.Println(err)
	}
}
