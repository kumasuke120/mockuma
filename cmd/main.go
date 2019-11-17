package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/kumasuke120/mockuma/internal/mapping"
	"github.com/kumasuke120/mockuma/internal/serve"
)

const defaultMapfile = "mockuMappings.json"

var port = flag.Int("p", 3214,
	"the port number on which Mockuma listen")
var mapfile = flag.String("mapfile", "mockuMappings.json",
	"the name of a json file which defines mockuMappings")

func init() {
	flag.Parse()

	if *mapfile != defaultMapfile {
		setWorkingDirBasedOnMapfile()
	}
}

func main() {
	mappings, err := mapping.FromJsonFile(*mapfile)
	if err != nil {
		log.Fatal("cannot load mockuMappings:", err)
	}

	server := &serve.MockServer{Port: *port, Mappings: mappings}
	if err := server.Start(); err != nil {
		log.Fatal("cannot start server:", err)
	}
}

func setWorkingDirBasedOnMapfile() {
	abs, err := filepath.Abs(*mapfile)
	if err != nil {
		log.Fatal("Cannot acquire the absolute path of mockuMappings:", err)
	}

	dir := filepath.Dir(abs)
	err = os.Chdir(dir)
	if err != nil {
		log.Fatal("Cannot change the working directory:", err)
	}

	log.Println("[main] working directory has been changed to:", dir)
}
