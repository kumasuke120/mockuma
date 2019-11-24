package mckmaps

import (
	"log"
	"os"
	"path/filepath"
)

var defaultMapfile = []string{
	"mockuMappings.json",
	"mockuMappings.main.json",
}

func LoadFromFile(filename string) (*MockuMappings, error) {
	return loadFromFile(filename, true)
}

func loadFromFile(filename string, chdir bool) (*MockuMappings, error) {
	var mappings *MockuMappings
	var err error

	if filename == "" {
		return loadFromDefaultMapfile(mappings, err)
	}

	filename, err = filepath.Abs(filename) // gets absolute path before chdir
	if err != nil {
		return nil, err
	}
	if chdir { // only changes once before any actual parsing
		if err = chdirBasedOnFilename(filename); err != nil {
			return nil, err
		}
	}

	parser := &parser{filename: filename}
	return parser.parse(chdir)
}

func loadFromDefaultMapfile(mappings *MockuMappings, err error) (*MockuMappings, error) {
	for _, _filename := range defaultMapfile {
		mappings, err = loadFromFile(_filename, false)
		if err == nil {
			return mappings, nil
		}
	}
	return nil, err
}

func chdirBasedOnFilename(filename string) error {
	dir := filepath.Dir(filename)
	err := os.Chdir(dir)
	if err != nil {
		return err
	}

	log.Println("[loader] working directory has been changed to:", dir)

	return nil
}
