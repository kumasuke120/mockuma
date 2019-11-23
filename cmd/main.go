package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/kumasuke120/mockuma/internal/mckmaps"
	"github.com/kumasuke120/mockuma/internal/server"
)

const (
	appName       = "MocKuma"
	versionNumber = "1.1.0"
)

var port = flag.Int("p", 3214,
	"sets the port number on which Mockuma listen")
var mapfile = flag.String("mapfile", "",
	"sets the name of a json file which defines mockuMappings")
var showVersion = flag.Bool("-version", false, "shows the version number")

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Printf("%s: %s\n", appName, versionNumber)
	} else {
		mappings, err := mckmaps.LoadFromJsonFile(*mapfile)
		if err != nil {
			log.Fatal("cannot load mockuMappings:", err)
		}

		s := server.NewMockServer(*port, mappings)
		s.SetServer(appName, versionNumber)
		if err := s.Start(); err != nil {
			log.Fatal("cannot start server:", err)
		}
	}
}
