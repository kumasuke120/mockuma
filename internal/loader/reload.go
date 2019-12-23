package loader

import (
	"log"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/kumasuke120/mockuma/internal/mckmaps"
)

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
				if onFileChange(filename, callback) {
					break watchLoop
				}
			}
		case wErr, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("[loader] failure encountered when watching files within the MockuMappings:", wErr)
		}
	}
}

func onFileChange(filename string, callback func(*mckmaps.MockuMappings)) bool {
	log.Println("[loader] at least one file within the MockuMappings was changed, reloading...")
	newMappings, err := LoadFromFile(filename)
	if err != nil {
		log.Println("[loader] fail to reload the MockuMappings:", err)
		return false
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

	return true
}
