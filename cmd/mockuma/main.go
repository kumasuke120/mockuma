//+build !test

package main

import (
	"flag"
	"log"
	"math/rand"
	"time"

	"github.com/kumasuke120/mockuma/internal"
	"github.com/kumasuke120/mockuma/internal/mckmaps"
	"github.com/kumasuke120/mockuma/internal/server"
)

var port = flag.Int("p", 3214,
	"sets the port number on which Mockuma listen")
var mapfile = flag.String("mapfile", "",
	"sets the name of a json file which defines mockuMappings")
var showVersion = flag.Bool("version", false, "shows the version information for MocKuma")

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	flag.Parse()

	if *showVersion {
		internal.PrintVersion()
	} else {
		mappings, err := mckmaps.LoadFromFile(*mapfile)
		if err != nil {
			log.Fatal("[main] cannot load mockuMappings:", err)
		}

		s := server.NewMockServer(*port, mappings)
		if err := s.Start(); err != nil {
			log.Fatal("[main] cannot start server:", err)
		}
	}
}
