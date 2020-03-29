package loader

import (
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/kumasuke120/mockuma/internal/mckmaps"
)

var defaultMapfile = []string{
	"mockuMappings.json",
	"mockuMappings.main.json",
	"main.json",
}

type Loader struct {
	mux      sync.Mutex
	filename string
	loaded   *mckmaps.MockuMappings
}

func New(filename string) *Loader {
	return &Loader{filename: filename}
}

func (l *Loader) Load() (*mckmaps.MockuMappings, error) {
	l.mux.Lock()
	defer l.mux.Unlock()

	if l.filename == "" {
		return l.loadDefault()
	}

	return l.loadFromFile(l.filename, true)
}

func (l *Loader) loadDefault() (m *mckmaps.MockuMappings, e error) {
	for _, f := range defaultMapfile {
		if _, e = os.Stat(f); !os.IsNotExist(e) {
			m, e = l.loadFromFile(f, false)
			return
		}
	}
	return
}

func (l *Loader) loadFromFile(filename string, chdir bool) (*mckmaps.MockuMappings, error) {
	var err error

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
	var mappings *mckmaps.MockuMappings
	mappings, err = parser.Parse()
	if err == nil { // saves filename if succeeded
		l.filename = filename
		l.loaded = mappings
	}
	return mappings, err
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

		log.Println("[loader  ] working directory has been changed to:", dir)
	}

	return nil
}
