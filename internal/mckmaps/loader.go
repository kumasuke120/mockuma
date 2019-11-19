package mckmaps

var defaultMapfile = []string{
	"mockuMappings.json",
	"mockuMappings.main.json",
}

func LoadFromJsonFile(filename string) (*MockuMappings, error) {
	return loadFromJsonFile(filename, true)
}

func loadFromJsonFile(filename string, chdir bool) (*MockuMappings, error) {
	var mappings *MockuMappings
	var err error

	if filename == "" {
		for _, _filename := range defaultMapfile {
			mappings, err = loadFromJsonFile(_filename, false)
			if err == nil {
				return mappings, nil
			}
		}

		return nil, err
	}

	parser := &parser{filename: filename, chdir: chdir}
	return parser.parse()
}
