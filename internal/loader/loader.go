package loader

import (
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/kumasuke120/mockuma/internal/mckmaps"
)

var defaultMapfile = []string{
	"mockuMappings.json",
	"mockuMappings.main.json",
}

func LoadFromFile(filename string) (*mckmaps.MockuMappings, error) {
	return loadFromFile(filename, true)
}

func loadFromFile(filename string, chdir bool) (*mckmaps.MockuMappings, error) {
	var mappings *mckmaps.MockuMappings
	var err error

	if filename == "" {
		return loadFromDefaultMapfile(mappings, err)
	}

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
	return parser.Parse()
}

func loadFromDefaultMapfile(mappings *mckmaps.MockuMappings, err error) (*mckmaps.MockuMappings, error) {
	for _, _filename := range defaultMapfile {
		mappings, err = loadFromFile(_filename, false)
		if err == nil {
			return mappings, nil
		}
	}
	return nil, err
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

		log.Println("[loader] working directory has been changed to:", dir)
	}

	return nil
}

var reloadMutex = &sync.Mutex{}

func EnableAutoReload(filenames []string, callback func(*mckmaps.MockuMappings)) error {
	if len(filenames) == 0 {
		panic("parameter 'filenames' should not be empty")
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	// adds filenames to watch
	for _, f := range filenames {
		err := watcher.Add(f)
		if err != nil {
			return err
		}
	}

	mainFilename := filenames[0] // the first filename should be the one of main file
	go watchFileChanges(watcher, mainFilename, callback)

	return nil
}

func watchFileChanges(watcher *fsnotify.Watcher, filename string, callback func(*mckmaps.MockuMappings)) {
	defer func() {
		if err := watcher.Close(); err != nil {
			log.Println("[loader] fail to close watcher:", err)
		}
	}()

watchLoop:
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			if event.Op&fsnotify.Write == fsnotify.Write {
				log.Println("[loader] at least one file within the MockuMappings was changed, reloading...")
				newMappings, err := LoadFromFile(filename)
				if err != nil {
					log.Println("[loader] fail to reload the MockuMappings:", err)
				}
				log.Println("[loader] a new MockuMappings was loaded")

				go func() {
					// there can be only one goroutine reload at the same time
					reloadMutex.Lock()
					defer reloadMutex.Unlock()

					// starts a new watcher goroutine, preventing from exiting
					if err = EnableAutoReload(newMappings.Filenames, callback); err != nil {
						log.Fatalln("[loader] fail to enable automatic reload:", err)
					}
					go callback(newMappings)
				}()
				break watchLoop
			}
		case wErr, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("[loader] failure encountered when watching files within the MockuMappings:", wErr)
		}
	}
}
