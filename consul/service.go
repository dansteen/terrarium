package consul

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/dansteen/terrarium/service"
	consul "github.com/hashicorp/consul/api"
	"github.com/rs/zerolog/log"
)

// Service is an instance of this service
type Service struct {
	service.Generic
}

// NewService will create a initialize an instance of the service with default values
func NewService(workspace string) *Service {
	// first initialize the generic stuff
	newService := Service{}
	newService.SetName("consul")
	// our config for the application. this is much easier than trying to work with hcl in a write context
	newService.SetServiceConfig(`
bootstrap_expect: 1
advertise_addr: "127.0.0.1"
client_addr: "127.0.0.1"
enable_syslog: true
ui: true
datacenter: "terrarium"
server: true
`)
	newService.SetWorkspace(workspace)
	newService.SetHealthyTimeout(30)
	newService.Version = "1.1.0"
	newService.ServiceConfigName = "consul_server.hcl"
	// hardcoded for now.  We need to change this to support multiple simultanious environments
	newService.Address = "127.0.0.1:8500"
	newService.Datadir = filepath.Join(workspace, newService.Name()+".d")
	newService.Logfile = filepath.Join(workspace, newService.Name()+".log")

	newService.Cmdline = fmt.Sprintf("%s agent -data-dir \"%s\" -config-file %s &> \"%s\"", filepath.Join(workspace, newService.Name()), newService.Datadir, newService.ServiceConfigName, newService.Logfile)
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
	client, err := consul.NewClient(&consul.Config{
		Address: service.Address,
		Scheme:  "http",
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
			// grab our health and return if its good
			health, err := client.Operator().AutopilotServerHealth(&consul.QueryOptions{})
			if err == nil {
				return health.Healthy, nil
			}
			// pause for a second to let things run
			time.Sleep(1 * time.Second)
		}
	}
	return false, errors.New("We shouldn't be here")
}
