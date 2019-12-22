//+build !test

package main

import (
	"flag"
	"log"
	"math/rand"
	"runtime"
	"time"

	"github.com/kumasuke120/mockuma/internal"
	"github.com/kumasuke120/mockuma/internal/loader"
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
		mappings, err := loader.LoadFromFile(*mapfile)
		if err != nil {
			log.Fatalln("[main] cannot load mockuMappings:", err)
		}

		s := server.NewMockServer(*port)
		if err = loader.EnableAutoReload(mappings.Filenames, s.SetMappings); err != nil {
			log.Fatalln("[main] fail to enable automatic reload:", err)
		}
		go s.Start(mappings)

		runtime.Goexit()
	}
}
