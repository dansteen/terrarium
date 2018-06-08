package command

import (
	"fmt"
	"os"

	"github.com/dansteen/terrarium/consul"
	"github.com/dansteen/terrarium/nomad"
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
	consulInstance, err := consul.GetService(workspace)
	if err != nil {
		log.Error().Err(err).Msgf("Unable to stop %s:", consulInstance.Name())
	} else {
		consulInstance.Stop()
	}

	vaultInstance, err := vault.GetService(workspace)
	if err != nil {
		log.Error().Err(err).Msgf("Unable to stop %s:", vaultInstance.Name())
	} else {
		vaultInstance.Stop()
	}

	nomadInstance, err := nomad.GetService(workspace)
	if err != nil {
		log.Error().Err(err).Msgf("Unable to stop %s:", nomadInstance.Name())
	} else {
		nomadInstance.Stop()
	}

}
