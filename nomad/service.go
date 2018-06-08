package nomad

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/dansteen/terrarium/service"
	nomad "github.com/hashicorp/nomad/api"
	"github.com/rs/zerolog/log"
)

// Service is an instance of this service
type Service struct {
	service.Generic
}

// NewService will create a initialize an instance of the service with default values
func NewService(workspace, consulAddress, vaultAddress, vaultToken string) *Service {
	// first initialize the generic stuff
	newService := Service{}
	newService.SetName("nomad")
	newService.SetWorkspace(workspace)
	newService.SetServiceConfig(fmt.Sprintf(`
datacenter = "terrarium"

server {
  enabled          = true
	bootstrap_expect = 1
	raft_protocol    = 3
}
client {
  enabled = true
}
consul {
  server_auto_join = true
  address          = "%s"
}
vault {
  enabled               = true
  token                 = "%s"
  address               = "%s"
  allow_unauthenticated = true
}`, consulAddress, vaultToken, vaultAddress))
	newService.SetHealthyTimeout(30)
	newService.Version = "0.8.3"
	newService.ServiceConfigName = "nomad_server.hcl"
	// hardcoded for now.  We need to change this to support multiple simultanious environments
	newService.Address = "http://127.0.0.1:4646"
	// our config for the application. this is much easier than trying to work with hcl in a write context
	newService.Datadir = filepath.Join(workspace, newService.Name()+".d")
	newService.Logfile = filepath.Join(workspace, newService.Name()+".log")

	newService.Cmdline = fmt.Sprintf("%s agent -data-dir \"%s\" -config \"%s\" &> \"%s\"", filepath.Join(workspace, newService.Name()), newService.Datadir, filepath.Join(newService.Datadir, newService.ServiceConfigName), newService.Logfile)
	newService.DownloadURL = fmt.Sprintf("https://releases.hashicorp.com/%s/%s/%s_%s_%s_%s.zip", newService.Name(), newService.Version, newService.Name(), newService.Version, runtime.GOOS, runtime.GOARCH)
	return &newService
}

// Healthy will check the health of the consul instance
func (service *Service) Healthy() (bool, error) {
	// first run our generic check
	healthy, err := service.Generic.Healthy()
	if err != nil {
		return healthy, err
	}

	// then check to see if it thinks its healthy
	client, err := nomad.NewClient(&nomad.Config{
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
			// we want to use this, but have to wait for a later version of nomad apparently
			//health, _, err := client.Operator().AutopilotServerHealth(&nomad.QueryOptions{})
			// in the meantime, we just see if we can get a leader out of nomad
			_, err := client.Status().Leader()
			if err == nil {
				return true, nil
			}
			// pause for a second to let things run
			time.Sleep(1 * time.Second)
		}
	}
	return false, errors.New("We shouldn't be here")
}
