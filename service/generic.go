package service

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	getter "github.com/hashicorp/go-getter"
	"github.com/rs/zerolog/log"
	yaml "gopkg.in/yaml.v2"
)

// Generic is a generic service intended to be overridden
type Generic struct {
	name              string `yaml:"name"`
	Cmdline           string `yaml:"cmdline"`
	Address           string `yaml:"address"`
	Pid               int    `yaml:"pid"`
	Version           string `yaml:"version"`
	Datadir           string `yaml:"datadir"`
	Logfile           string `yaml:"logfile"`
	DownloadURL       string `yaml:"download_url"`
	ServiceConfigName string `yaml:"service_config_name"`
	healthyTimeout    int    `yaml:"healthy_timeout"`
	serviceConfig     string
	workspace         string
}

// Init will generate a new service for this workspace unless one already exists
// and will return true if it generates one and false if it exists
func (service *Generic) Init() error {

	// Make sure we have the binary we need
	if _, err := os.Stat(filepath.Join(service.Workspace(), service.Name())); err != nil {
		log.Info().Msgf("Existing %s binary not found", strings.Title(service.Name()))
		err := service.Download()
		if err != nil {
			return err
		}
	}

	// create our datadir if it does not exist
	if _, err := os.Stat(service.Datadir); err != nil {
		err = os.Mkdir(service.Datadir, 0755)
		if err != nil {
			log.Error().Err(err).Msgf("Could not create %s data dir:", service.Name())
			return err
		}
	}

	return nil

}

// Download will download the app to our environment
func (service *Generic) Download() error {
	destination := filepath.Join(service.Workspace(), service.Name())
	log.Info().Msgf("Downloading from %s...", service.DownloadURL)
	err := getter.GetFile(destination, service.DownloadURL)
	if err != nil {
		log.Error().Err(err).Msgf("Could not download %s from %s:", strings.Title(service.Name()), service.DownloadURL)
	}
	return err
}

// WriteServiceConfig will create the config for the application we are starting
func (service *Generic) WriteServiceConfig() error {
	// the location of the config file
	configPath := filepath.Join(service.Datadir, service.ServiceConfigName)
	// in the future we will need to instantiate this as a template, but for now this is fine
	data := []byte(service.serviceConfig)

	// and write it out
	err := ioutil.WriteFile(configPath, data, 0644)
	if err != nil {
		log.Error().Err(err).Msgf("Error writing %s config data to %s", service.Name(), configPath)
		return err
	}
	return nil
}

// Read will read an existing instance from the services workspace
func (service *Generic) Read() (bool, error) {
	// the location of the config file
	configPath := filepath.Join(service.Workspace(), service.Name()+".yml")

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

// Write will write instance data to a file in workspace named <app>.yml
func (service *Generic) Write() error {
	configPath := filepath.Join(service.Workspace(), service.Name()+".yml")

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

// Healthy will check the health of the service.  This will only check if the process exists. More advanced healthchecks
// must be implemented by each SupportService.
func (service *Generic) Healthy() (bool, error) {
	// the first check is to see if the process is running
	process, _ := os.FindProcess(service.Pid)
	err := process.Signal(os.Signal(syscall.Signal(0x0)))
	if err != nil {
		return false, err
	}
	return true, nil
}

// Workspace will return the workspace that this service runs in
func (service *Generic) Workspace() string {
	return service.workspace
}

// SetWorkspace will return the workspace that this service runs in
func (service *Generic) SetWorkspace(workspace string) {
	service.workspace = workspace
}

// Name will rturn the name of the service
func (service *Generic) Name() string {
	return service.name
}

// SetName will return the workspace that this service runs in
func (service *Generic) SetName(name string) {
	service.name = name
}

// ServiceConfig will return the service config for this service
func (service *Generic) ServiceConfig() string {
	return service.serviceConfig
}

// SetServiceConfig will return the workspace that this service runs in
func (service *Generic) SetServiceConfig(serviceConfig string) {
	service.serviceConfig = serviceConfig
}

// HealthyTimeout will return the healthy timeout for this service
func (service *Generic) HealthyTimeout() int {
	return service.healthyTimeout
}

// SetHealthyTimeout will return the workspace that this service runs in
func (service *Generic) SetHealthyTimeout(timeout int) {
	service.healthyTimeout = timeout
}

// Start will start consul for this environemnt
func (service *Generic) Start() error {
	log.Info().Msgf("Starting %s", service.Name())
	// start up our command
	cmd := exec.Command("bash", "-c", fmt.Sprintf("exec %s", service.Cmdline))
	err := cmd.Start()
	if err != nil {
		log.Error().Err(err)
		return err
	}
	// save off some values
	service.Pid = cmd.Process.Pid
	// once it comes up write our config
	err = service.Write()
	if err != nil {
		// stop the service so we dont leave running processes around
		service.Stop()
		return err
	}

	log.Info().Msgf("Started")
	return nil
}

// Stop will stop consul for this environment
func (service *Generic) Stop() error {
	log.Info().Msgf("Stopping %s", strings.Title(service.Name()))
	// send a kill signal to the process
	process, _ := os.FindProcess(service.Pid)
	err := process.Signal(os.Signal(os.Interrupt))
	if err != nil && err.Error() != "os: process already finished" && err.Error() != "os: process not initialized" {
		log.Error().Err(err).Msgf("Could not stop %s (pid %s)", strings.Title(service.Name()), service.Pid)
		return err
	}

	log.Info().Msg("Stopped")
	return nil
}

// Restart will stop the service for this environment and then restart it.
func (service *Generic) Restart() error {
	err := service.Stop()
	if err != nil {
		return err
	}

	service.Start()
	if err != nil {
		return err
	}
	return nil
}
