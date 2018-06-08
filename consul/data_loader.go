package consul

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/dansteen/terrarium/utility"
	consul "github.com/hashicorp/consul/api"
	"github.com/rs/zerolog/log"
	yaml "gopkg.in/yaml.v2"
)

// Load will accept a path to a yaml file and will load the content of that file into consul using the methodology described here:
// https://github.com/traitify/ops_scripts/blob/master/CONSUL_ORGANIZATION.md#app
func (service *Service) Load(dataFile, appName, hashLabel string) error {

	// make sure the file exists
	if _, err := os.Stat(dataFile); err != nil {
		log.Warn().Msg("No data file found. Skipping.")
		return nil
	}

	// read the file
	content, err := ioutil.ReadFile(dataFile)
	if err != nil {
		log.Error().Err(err).Msgf("Error reading config file at %s.", dataFile)
		return err
	}
	// create our data structure
	data := utility.YamlData{DataType: "consul"}
	yaml.Unmarshal(content, &data)

	// create a consul connection
	client, err := consul.NewClient(&consul.Config{
		Address: service.Address,
		Scheme:  "http",
	})
	if err != nil {
		log.Error().Err(err)
		return err
	}

	// run through our records and create keys
	for key, value := range data.Records {
		_, err = client.KV().Put(&consul.KVPair{Key: filepath.Join("app", appName, hashLabel, key), Value: []byte(value)}, &consul.WriteOptions{})
		if err != nil {
			log.Error().Err(err).Msgf("Could not load application data file %s to consul:", dataFile)
			return err
		}
	}

	log.Info().Msgf("Loaded data file %s into consul", dataFile)
	return nil
}
