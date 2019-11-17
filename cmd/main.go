package main

import (
	"flag"
	"log"

	"github.com/kumasuke120/mockuma/internal/mapping"
	"github.com/kumasuke120/mockuma/internal/serve"
)

var port = flag.Int("p", 3214,
	"the port number on which Mockuma listen")
var mapfile = flag.String("mapfile", "mockuMappings.json",
	"the name of a json file which defines mockuMappings")

func main() {
	flag.Parse()

	mappings, err := mapping.FromJsonFile(*mapfile)
	if err != nil {
		log.Fatal("Cannot load mockuMappings: ", err)
	}

	server := &serve.MockServer{Port: *port, Mappings: mappings}
	if err := server.Start(); err != nil {
		log.Fatal("Cannot start server: ", err)
	}
}
