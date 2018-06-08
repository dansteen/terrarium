package consul

import (
	"io/ioutil"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/rs/zerolog/log"
	yaml "gopkg.in/yaml.v2"
)

// Data stores a representation of our consul data in a fashion that it can be easily added to consul
type Data struct {
	hashLabel string
	Records   map[string]string
}

// UnmarshalYAML(unmarshal func(interface{}) error) error

// UnmarshalYAML allows for custom unmarshaling
func (data Data) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var rawData interface{}
	// parse the data into a variable
	unmarshal(rawData)
	// run through and add data to our struct
	spew.Dump(rawData)
	return nil
}

// Load will accept a path to a yaml file and will load the content of that file into consul using the methodology described here:
// https://github.com/traitify/ops_scripts/blob/master/CONSUL_ORGANIZATION.md#app
func Load(dataFile string, hashLabel string) error {

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
	data := Data{hashLabel: hashLabel}
	yaml.Unmarshal(content, &data)

	return nil
}
