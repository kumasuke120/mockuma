package main

import (
	"flag"
	"log"

	"github.com/kumasuke120/mockuma/mapping"
	"github.com/kumasuke120/mockuma/serve"
)

func main() {
	port, mappingsFilename := parseAndGetFlags()

	mappings, err := mapping.FromJsonFile(mappingsFilename)
	if err != nil {
		log.Fatal("Cannot load mockuMappings: ", err)
	}

	server := &serve.MockServer{Port: port, Mappings: mappings}
	if err := server.Start(); err != nil {
		log.Fatal("Cannot start server: ", err)
	}
}

func parseAndGetFlags() (int, string) {
	port := flag.Int("p", 3214,
		"the port number on which Mockuma listen")
	mappingsFilename := flag.String("mapfile", "mockuMappings.json",
		"the name of a json file which defines mockuMappings")
	flag.Parse()
	return *port, *mappingsFilename
}
