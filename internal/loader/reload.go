package loader

import (
	"log"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/kumasuke120/mockuma/internal/mckmaps"
)

type fileChangeListener interface {
	onFileChange(path string)
}

type wdWatcher struct {
	wd        string
	watcher   *fsnotify.Watcher
	filenames []string
	listeners []fileChangeListener
	watching  *int32
}

func newWatcher(filenames []string) (*wdWatcher, error) {
	if len(filenames) == 0 {
		panic("parameter 'filenames' should not be empty")
	} else if anyAbs(filenames) {
		panic("parameter 'filenames' shouldn't contains absolute path")
	}

	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	watcher := &wdWatcher{
		wd:        wd,
		watcher:   fsWatcher,
		filenames: filenames,
		watching:  new(int32),
	}

	// watches the current working directory
	err = watcher.addWatchRecursively(wd)
	if err != nil {
		return nil, err
	}

	return watcher, nil
}

func anyAbs(names []string) bool {
	for _, n := range names {
		if filepath.IsAbs(n) {
			return true
		}
	}
	return false
}

func (w *wdWatcher) addWatchRecursively(name string) error {
	if !filepath.IsAbs(name) { // ensures absolute path
		if abs, err := filepath.Abs(name); err == nil {
			name = abs
		} else {
			return err
		}
	}

	if s, err := os.Stat(name); err == nil { // only adds if name represents a directory
		if !s.IsDir() {
			return nil
		}
	} else {
		return err
	}

	err := w.watcher.Add(name)
	if err != nil {
		return err
	}

	return filepath.Walk(name,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if path == name {
				return nil
			}

			if info.IsDir() {
				return w.addWatchRecursively(path)
			}
			return nil
		})
}

func (w *wdWatcher) addListener(listener fileChangeListener) {
	w.listeners = append(w.listeners, listener)
}

func (w *wdWatcher) watch() {
	defer func() {
		if err := w.watcher.Close(); err != nil {
			log.Println("[loader] fail to close watcher:", err)
		}
	}()

	atomic.StoreInt32(w.watching, 1)
	for atomic.LoadInt32(w.watching) == 1 { // w.watching is a flag indicates if keep watching or not
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}

			name := event.Name
			if event.Op&fsnotify.Create == fsnotify.Create {
				if err := w.addWatchRecursively(name); err != nil {
					log.Fatalln("[loader] fail to enable automatic reloading:", err)
				}
			}
			if w.isConcernedFile(name) {
				w.notifyAll(name)
			}
		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			log.Println("[loader] failure encountered when watching files:", err)
		default:
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func (w *wdWatcher) isConcernedFile(name string) bool {
	for _, f := range w.filenames {
		af := filepath.Join(w.wd, f)
		match, err := filepath.Match(af, name)
		if err != nil {
			panic("Shouldn't happen")
		}
		if match {
			return true
		}
	}
	return false
}

func (w *wdWatcher) notifyAll(name string) {
	for _, l := range w.listeners {
		l.onFileChange(name)
	}
}

func (w *wdWatcher) cancel() {
	atomic.StoreInt32(w.watching, 0)
}

func (l *Loader) EnableAutoReload(callback func(mappings *mckmaps.MockuMappings)) error {
	l.mux.Lock()
	defer l.mux.Unlock()

	if l.loaded == nil {
		panic("mockuMappings has not been loaded")
	}

	watcher, err := newWatcher(l.loaded.Filenames)
	if err != nil {
		return err
	}
	listener := &autoReloadListener{
		l:        l,
		w:        watcher,
		callback: callback,
	}
	watcher.addListener(listener)
	go watcher.watch()

	return nil
}

type autoReloadListener struct {
	l         *Loader
	w         *wdWatcher
	callback  func(mappings *mckmaps.MockuMappings)
	reloadMux sync.Mutex
}

func (l *autoReloadListener) onFileChange(path string) {
	log.Println("[loader] change detected:", path)

	mappings, err := l.l.Load()
	if err != nil {
		log.Println("[loader] cannot load mockuMappings after changing:", err)
		return
	}

	// there can be only one goroutine reloading at the same time
	l.reloadMux.Lock()
	defer l.reloadMux.Unlock()

	// starts a new watcher goroutine, preventing from exiting
	if err := l.l.EnableAutoReload(l.callback); err != nil {
		log.Fatalln("[loader] fail to enable automatic reloading:", err)
	}
	go l.callback(mappings)

	l.w.cancel()
}
