package command

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dansteen/terrarium/consul"
	"github.com/dansteen/terrarium/nomad"
	"github.com/dansteen/terrarium/service"
	"github.com/dansteen/terrarium/vault"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Shutdown will initialize this terrarium environment by creating the workspace and starting the support applications
func Shutdown(cmd *cobra.Command, args []string) {
	// grab our workspace
	workspace := viper.GetString("workspace")

	// check to see if it exists
	if _, err := os.Stat(fmt.Sprintf("%s", workspace)); err != nil {
		log.Error().Err(err).Msgf("Could not find workspace %s: ", workspace)
		os.Exit(1)
	}

	// spin down our applications
	consulInstance := consul.NewService(workspace)
	stopService(consulInstance)

	vaultInstance, _ := vault.NewService(workspace)
	stopService(vaultInstance)

	nomadInstance := nomad.NewService(workspace, consulInstance.Address, vaultInstance.Address, vaultInstance.RootToken)
	stopService(nomadInstance)

}

// stopService will start up a support service or restart it if its unhealthy
func stopService(service service.SupportService) error {

	configPath := filepath.Join(service.Workspace(), service.Name()+".yml")
	// first see if we have an existing config
	read, err := service.Read(configPath)
	if err != nil {
		log.Error().Err(err).Msgf("Unable to stop %s:", service.Name())
		return err
	}
	if !read {
		log.Error().Msgf("Unable to stop %s - no config found", service.Name())
		return nil
	}

	// if we do find a config we stop the service
	err = service.Stop()
	return err
}
