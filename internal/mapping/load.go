package mapping

import (
	"io/ioutil"
)

func FromJsonFile(filename string) (*MockuMappings, error) {
	mappingsJson, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return parseFromJson(mappingsJson)
}
