// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
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
	"github.com/dansteen/terrarium/command"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start an application in this environment",
	Long:  `Start an application in the environment`,
	Run:   command.StartApp,
}

func init() {
	rootCmd.AddCommand(startCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	startCmd.PersistentFlags().StringP("appPath", "a", "", "Path to the application directory to add")
	startCmd.PersistentFlags().StringP("hashLabel", "l", "", "arbitrary version label to use for this application")
	startCmd.MarkFlagRequired("appPath")
	startCmd.MarkFlagRequired("hashLabel")
	viper.BindPFlag("appPath", startCmd.PersistentFlags().Lookup("appPath"))
	viper.BindPFlag("hashLabel", startCmd.PersistentFlags().Lookup("hashLabel"))

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// startCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
