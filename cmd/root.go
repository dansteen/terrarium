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
	"fmt"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "terrarium",
	Short: "Build a development environment for local development.",
	Long: `Terrarium builds a and maintains a development environemnt
	using consul, vault, and nomad.   It will allow you to initialize an
	environment, and start and stop applications inside of that environment.
	It also manages the data for those applications in consul and vault.
	`,
	// we need to load in our project name
	//Args:  cobra.ExactArgs(1),
	//PersistentPreRun: loadProject,
	Run: func(cmd *cobra.Command, args []string) { cmd.Help() },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.terrarium.yaml)")
	rootCmd.PersistentFlags().StringP("project", "p", "default", "the name of the project")

	// set the workdir from our project name
	viper.BindPFlag("project", rootCmd.PersistentFlags().Lookup("project"))
	viper.Set("workspace", fmt.Sprintf("/tmp/terrarium_%s", rootCmd.PersistentFlags().Lookup("project").Value))

	// we are running in the console, so we use the console logger
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".terrarium" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".terrarium")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

// loadProject will load up the name of the project you want to operate on
func loadProject(cmd *cobra.Command, args []string) {

	// we are garanteed to have an args[0]
	viper.Set("project", os.Args[1])
	fmt.Println(viper.Get("project"))
}
