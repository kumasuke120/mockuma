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
	mux          sync.Mutex
	oldWd        string
	filename     string
	loadFilename string
	watcher      *fileWatcher
	loaded       *mckmaps.MockuMappings
	zipMode      bool
	tempDirs     []string
}

func New(filename string) *Loader {
	wd := myos.GetWd()
	return &Loader{filename: filename, zipMode: false, oldWd: wd}
}

func (l *Loader) Load() (*mckmaps.MockuMappings, error) {
	l.mux.Lock()
	defer l.mux.Unlock()

	err := l.absFilename() // gets absolute path for chdir and fsnotify
	if err != nil {
		return nil, err
	}

	l.zipMode = filepath.Ext(l.filename) == ".zip"
	if l.zipMode {
		err = l.beforeLoadZip()
	} else {
		err = l.beforeLoadNormal()
	}
	if err != nil {
		return nil, err
	}

	return l.loadFromFile(l.loadFilename)
}

func (l *Loader) absFilename() error {
	if l.filename == "" {
		return nil
	}

	_filename, err := filepath.Abs(l.filename)
	if err != nil {
		return err
	}
	l.filename = _filename

	return nil
}

func (l *Loader) beforeLoadZip() error {
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

	def, err := l.getExistentDefault()
	if err != nil {
		return err
	}
	l.loadFilename = def

	return nil
}

func (l *Loader) beforeLoadNormal() error {
	if l.filename == "" {
		def, err := l.getExistentDefault()
		if err != nil {
			return err
		}
		l.loadFilename = def
	} else {
		l.loadFilename = l.filename
	}

	// only changes once before any actual parsing
	if err := chdirBasedOnFilename(l.loadFilename); err != nil {
		return err
	}

	return nil
}

func (l *Loader) getExistentDefault() (string, error) {
	var err error
	for _, f := range defaultMapfile {
		if _, err = os.Stat(f); !os.IsNotExist(err) { // if file exists
			f, err = filepath.Abs(f)
			return f, err
		}
	}
	return "", err
}

func (l *Loader) loadFromFile(filename string) (*mckmaps.MockuMappings, error) {
	parser := mckmaps.NewParser(filename)
	mappings, err := parser.Parse()
	if err == nil { // saves loaded mappings if succeeded
		l.loaded = mappings
	}
	return mappings, err
}

func (l *Loader) Clean() error {
	if len(l.tempDirs) == 0 {
		return nil
	}

	// releases the directory for removing
	if l.watcher != nil {
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
