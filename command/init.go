package command

import (
	"fmt"
	"os"
	"strings"

	"github.com/dansteen/terrarium/consul"
	"github.com/dansteen/terrarium/nomad"
	"github.com/dansteen/terrarium/service"
	"github.com/dansteen/terrarium/vault"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// InitEnv will initialize this terrarium environment by creating the workspace and starting the support applications
func InitEnv(cmd *cobra.Command, args []string) {
	// grab our workspace
	workspace := viper.GetString("workspace")
	// check to see if it exists
	if _, err := os.Stat(fmt.Sprintf("%s", workspace)); err == nil {
		log.Info().Msg("Found existing project. Health Checking.")
	} else {
		log.Info().Msgf("Creating project at %s.", workspace)
		err := os.Mkdir(workspace, 0755)
		if err != nil {
			log.Error().Err(err).Msgf("Could not create project")
			os.Exit(1)
		}
	}

	// spin up consul
	consulInstance, err := consul.NewService(workspace)
	if err != nil {
		os.Exit(1)
	}
	err = startService(consulInstance)
	if err != nil {
		os.Exit(1)
	}

	// spin up vault
	vaultInstance, err := vault.NewService(workspace)
	if err != nil {
		os.Exit(1)
	}
	err = startService(vaultInstance)
	if err != nil {
		os.Exit(1)
	}
	// configure our vault backents
	err = vaultInstance.ConfigureBackends()
	if err != nil {
		os.Exit(1)
	}

	// spin up nomad
	nomadInstance, err := nomad.NewService(workspace, consulInstance.Address, vaultInstance.Address, vaultInstance.RootToken)
	if err != nil {
		os.Exit(1)
	}
	err = startService(nomadInstance)
	if err != nil {
		os.Exit(1)
	}
}

// startService will start up a support service or restart it if its unhealthy
func startService(service service.SupportService) error {

	// first see if we have an existing config in the services workspace
	read, err := service.Read()
	if err != nil {
		return err
	}

	err = service.Init()
	if err != nil {
		os.Exit(1)
	}
	// if there is already an instance in this workspace
	if read {
		log.Info().Msgf("Existing %s Instance found. Checking...", strings.Title(service.Name()))
		// check to see if its healthy
		healthy, _ := service.Healthy()
		// if we are healthy we return
		if healthy {
			log.Info().Msgf("%s is Healthy.", strings.Title(service.Name()))
			return nil
		}

		// depending on the error returned we do things a bit differently
		// if the process does not exist
		//if err.Error() == "os: process not initialized" || err.Error() == "os: process already finished" {
		//		log.Warn().Msgf("%s is not running.", strings.Title(service.Name()))
		//	} else {
		log.Warn().Msgf("%s is not healthy. Restarting...", strings.Title(service.Name()))
		//}
		service.Stop()
	}

	// regardless we issue a start
	err = service.Start()
	if err != nil {
		return err
	}

	// then, we write the service config
	err = service.WriteServiceConfig()
	if err != nil {
		return err
	}
	// and and our config to make sure we have all the information we need
	err = service.Write()
	if err != nil {
		return err
	}

	// then we make sure things are healthy
	log.Info().Msgf("Waiting %d seconds for %s to come up", service.HealthyTimeout(), strings.Title(service.Name()))
	healthy, err := service.Healthy()
	if err != nil || !healthy {
		// if we had an error we stop the process
		if err.Error() == "os: process not initialized" {
			log.Error().Err(err).Msgf("Could not start %s", strings.Title(service.Name()))
		} else {
			log.Error().Err(err).Msgf("%s not Healthy", strings.Title(service.Name()))
		}
		service.Stop()
		return err
	}
	return nil
}
