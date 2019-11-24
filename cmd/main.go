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
	author        = "kumasuke120<bearcomingx@gmail.com>"
	github        = "https://github.com/kumasuke120/mockuma"
)

var port = flag.Int("p", 3214,
	"sets the port number on which Mockuma listen")
var mapfile = flag.String("mapfile", "",
	"sets the name of a json file which defines mockuMappings")
var showVersion = flag.Bool("version", false, "shows the version information for MocKuma")

func main() {
	flag.Parse()

	if *showVersion {
		printVersion()
	} else {
		mappings, err := mckmaps.LoadFromFile(*mapfile)
		if err != nil {
			log.Fatal("[main] cannot load mockuMappings: ", err)
		}

		s := server.NewMockServer(*port, mappings)
		s.SetNameAndVersion(appName, versionNumber)
		if err := s.Start(); err != nil {
			log.Fatal("[main] cannot start server: ", err)
		}
	}
}

func printVersion() {
	fmt.Println(` _______              __  __                       `)
	fmt.Println(`|   |   |.-----.----.|  |/  |.--.--.--------.---.-.`)
	fmt.Println(`|       ||  _  |  __||     < |  |  |        |  _  |`)
	fmt.Println(`|__|_|__||_____|____||__|\__||_____|__|__|__|___._|`)
	fmt.Println()
	fmt.Printf("Version\t: %s\n", versionNumber)
	fmt.Printf("Author\t: %s\n", author)
	fmt.Printf("GitHub\t: %s\n", github)
}
