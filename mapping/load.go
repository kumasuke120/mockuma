package mapping

import (
	"encoding/json"
	"io/ioutil"
)

func FromJsonFile(filename string) (*MockuMappings, error) {
	mappingsJson, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	mappings, err := parseFromJson(mappingsJson)
	if err != nil {
		return nil, err
	}

	return &MockuMappings{mappings: mappings}, nil
}

func parseFromJson(jsonData []byte) (map[string][]*MockuMapping, error) {
	var v interface{}
	err := json.Unmarshal(jsonData, &v)
	if err != nil {
		return nil, err
	}

	data := v.([]interface{})
	return parseAsMockuMappingMap(data)
}
