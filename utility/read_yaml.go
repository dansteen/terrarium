package utility

import (
	"strings"
)

// YamlData stores a representation of our config data in a fashion that it can be easily added to consul or vault
type YamlData struct {
	Records map[string]string
	// either consul or vault.  This impacts how we format the data
	DataType string
}

// UnmarshalYAML allows for custom unmarshaling based on the destination of the data
func (data *YamlData) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var rawData interface{}
	var newKey string
	// create a new map
	data.Records = make(map[string]string)

	// parse the data into a variable
	unmarshal(&rawData)
	// flatten our data
	flatmap := Flatten(rawData.(map[interface{}]interface{}))
	// run through and adjust our data ppropriately for the application
	switch data.DataType {
	case "consul":
		for key, value := range flatmap {
			// adjust the key to match our conventions
			keyParts := strings.SplitN(key, "/", 2)
			if keyParts[0] == "default" {
				newKey = keyParts[1]
			} else {
				newKey = keyParts[1] + "/" + keyParts[0]
			}
			// then put the value in our struct
			data.Records[newKey] = value
		}
	case "vault":
		data.Records = flatmap
	}

	return nil
}
