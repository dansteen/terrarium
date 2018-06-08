package vault

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	yaml "gopkg.in/yaml.v2"

	"github.com/dansteen/terrarium/service"
	vault "github.com/hashicorp/vault/api"
	"github.com/rs/zerolog/log"
	"github.com/satori/go.uuid"
)

// Service is an instance of this service
type Service struct {
	service.Generic
	RootToken string `yaml:"root_token"`
}

// NewService will create a initialize an instance of the service with default values
func NewService(workspace string) (*Service, error) {
	// first initialize the generic stuff
	newService := Service{}
	newService.SetName("vault")
	newService.SetWorkspace(workspace)
	newService.SetServiceConfig("")
	newService.SetHealthyTimeout(30)
	newService.Version = "0.10.1"
	newService.ServiceConfigName = "vault_server.hcl"
	// hardcoded for now.  We need to change this to support multiple simultanious environments
	newService.Address = "http://127.0.0.1:8200"
	newService.Datadir = filepath.Join(workspace, newService.Name()+".d")
	newService.Logfile = filepath.Join(workspace, newService.Name()+".log")

	// generate a root token
	rootToken, err := uuid.NewV4()
	if err != nil {
		log.Error().Err(err).Msg("Could not generate a root token for vault:")
		return &newService, err
	}
	newService.RootToken = rootToken.String()

	newService.Cmdline = fmt.Sprintf("%s server -dev -dev-root-token-id %s &> \"%s\"", filepath.Join(workspace, newService.Name()), newService.RootToken, newService.Logfile)
	newService.DownloadURL = fmt.Sprintf("https://releases.hashicorp.com/%s/%s/%s_%s_%s_%s.zip", newService.Name(), newService.Version, newService.Name(), newService.Version, runtime.GOOS, runtime.GOARCH)
	return &newService, nil
}

// Healthy will check the health of the consul instance
func (service *Service) Healthy() (bool, error) {
	// first run our generic check
	healthy, err := service.Generic.Healthy()
	if err != nil {
		return healthy, err
	}

	// then check to see if it thinks its healthy
	client, err := vault.NewClient(&vault.Config{
		Address: service.Address,
	})
	if err != nil {
		log.Error().Err(err)
		return false, err
	}

	// set up our timer
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(30 * time.Second)
		timeout <- false
	}()

	// keep this up until things return or we timeout
	for true {
		select {
		// we give things 30 seconds to come up
		case <-timeout:
			err = fmt.Errorf("Timeout exceeded while starting %s", strings.Title(service.Name()))
			log.Error().Err(err)
			return false, err
		default:
			// grab our health and return depending on the value
			health, err := client.Sys().Health()
			if err == nil {
				if health.Sealed {
					err = errors.New("Vault is inexplicably sealed")
					log.Error().Err(err)
					return false, err
				}
				if health.Standby {
					err = errors.New("Vault is inexplicably in standby mode")
					log.Error().Err(err)
					return false, err
				}
				if !health.Initialized {
					err = errors.New("Vault is not initialized")
					log.Error().Err(err)
					return false, err
				}
				// if we get here things are good
				return true, nil
			}
			// pause for a second to let things run
			time.Sleep(1 * time.Second)
		}
	}
	return false, errors.New("We shouldn't be here")
}

// Read will read an existing instance.  We need to overide the generic reader here to ensure that we get our extra stanzas
func (service *Service) Read(configPath string) (bool, error) {

	// if there is existing instance data
	if _, err := os.Stat(configPath); err == nil {
		// if there is read it in (these files are short so we can read the whole thing)
		content, err := ioutil.ReadFile(configPath)
		if err != nil {
			fmt.Printf("%v", err)
			log.Error().Err(err).Msgf("Error reading config file at %s.", configPath)
			return false, err
		}

		err = yaml.Unmarshal(content, service)
		if err != nil {
			log.Error().Err(err).Msgf("Error processing config file content: %s.", configPath)
			return false, err
		}
		return true, nil
	}
	// we return false if there is no config file to read
	return false, nil
}

// Write will write instance data to a file in workspace named <app>.yml.   We need to overide the generic reader here to ensure that we get our extra stanzas
func (service *Service) Write(configPath string) error {

	// marshall our data
	data, err := yaml.Marshal(service)
	if err != nil {
		log.Error().Err(err).Msg("Error processing config data for writing")
		return err
	}
	// and write it out
	err = ioutil.WriteFile(configPath, data, 0644)
	if err != nil {
		log.Error().Err(err).Msgf("Error writing config data to %s", configPath)
		return err
	}
	return nil
}
