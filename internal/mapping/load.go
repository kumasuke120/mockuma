package mapping

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

const DefaultMapfile = "mockuMappings.json"

func FromJsonFile(filename string) (*MockuMappings, error) {
	mappingsJson, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	if filename != DefaultMapfile {
		setWorkingDirBasedOnFilename(filename)
	}

	return parseFromJson(mappingsJson)
}

func setWorkingDirBasedOnFilename(filename string) {
	abs, err := filepath.Abs(filename)
	if err != nil {
		log.Fatal("Cannot acquire the absolute path of mockuMappings:", err)
	}

	dir := filepath.Dir(abs)
	err = os.Chdir(dir)
	if err != nil {
		log.Fatal("Cannot change the working directory:", err)
	}

	log.Println("[load] working directory has been changed to:", dir)
}
