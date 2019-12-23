package loader

import (
	"log"
	"os"
	"path/filepath"

	"github.com/kumasuke120/mockuma/internal/mckmaps"
)

var defaultMapfile = []string{
	"mockuMappings.json",
	"mockuMappings.main.json",
}

func LoadFromFile(filename string) (*mckmaps.MockuMappings, error) {
	return loadFromFile(filename, true)
}

func loadFromFile(filename string, chdir bool) (*mckmaps.MockuMappings, error) {
	var mappings *mckmaps.MockuMappings
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

	parser := mckmaps.NewParser(filename)
	return parser.Parse()
}

func loadFromDefaultMapfile(mappings *mckmaps.MockuMappings, err error) (*mckmaps.MockuMappings, error) {
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
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	if wd != dir {
		err = os.Chdir(dir)
		if err != nil {
			return err
		}

		log.Println("[loader] working directory has been changed to:", dir)
	}

	return nil
}
