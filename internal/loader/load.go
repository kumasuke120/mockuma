package loader

import (
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/kumasuke120/mockuma/internal/mckmaps"
	"github.com/kumasuke120/mockuma/internal/myos"
)

var defaultMapfile = []string{
	"mockuMappings.json",
	"mockuMappings.main.json",
	"main.json",
}

type Loader struct {
	mux      sync.Mutex
	oldWd    string
	filename string
	watcher  *fileWatcher
	loaded   *mckmaps.MockuMappings
	zipMode  bool
	tempDirs []string
}

func New(filename string) *Loader {
	wd := myos.GetWd()
	return &Loader{filename: filename, zipMode: false, oldWd: wd}
}

func (l *Loader) Load() (*mckmaps.MockuMappings, error) {
	l.mux.Lock()
	defer l.mux.Unlock()

	if l.filename == "" {
		return l.loadDefault()
	} else if filepath.Ext(l.filename) == ".zip" {
		err := l.beforeLoadZip()
		if err != nil {
			return nil, err
		}
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

func (l *Loader) beforeLoadZip() error {
	l.zipMode = true

	err := l.Clean()
	if err != nil {
		log.Println("[loader  ] fail to clean temporary directories: " + err.Error())
	}

	dir, err := unzip(l.filename)
	if err != nil {
		return err
	}
	l.tempDirs = append(l.tempDirs, dir)

	err = chdirBasedOnDir(dir)
	if err != nil {
		return err
	}

	return nil
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
	if err == nil { // saves loaded mappings if succeeded
		l.loaded = mappings
	}
	return mappings, err
}

func (l *Loader) Clean() error {
	if len(l.tempDirs) == 0 {
		return nil
	}

	// releases the directory and watcher for removing
	if l.watcher != nil {
		l.watcher.cancel()

		err := myos.Chdir(l.oldWd)
		if err != nil {
			return err
		}
	}

	// deletes all temporary directories
	for _, dir := range l.tempDirs {
		err := os.RemoveAll(dir)
		if err != nil {
			return err
		}
	}
	l.tempDirs = nil

	return nil
}

func chdirBasedOnFilename(filename string) error {
	dir := filepath.Dir(filename)
	return chdirBasedOnDir(dir)
}

func chdirBasedOnDir(dir string) error {
	wd := myos.GetWd()

	if wd != dir {
		err := myos.Chdir(dir)
		if err != nil {
			return err
		}

		log.Println("[loader  ] chdir    : working directory changed:", dir)
	}

	return nil
}
