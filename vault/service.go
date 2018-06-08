package vault

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"time"

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
