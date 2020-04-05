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
	"github.com/kumasuke120/mockuma/internal/myos"
)

const watchInterval = 50 * time.Millisecond

type fileChangeListener interface {
	onFileChange(path string)
}

type fileWatcher struct {
	wd        string
	watcher   *fsnotify.Watcher
	filenames []string
	listeners []fileChangeListener
	watching  *int32
}

func newWdWatcher(filenames []string) (*fileWatcher, error) {
	if len(filenames) == 0 {
		panic("parameter 'filenames' should not be empty")
	} else if anyAbs(filenames) {
		panic("parameter 'filenames' shouldn't contains absolute path")
	}

	wd := myos.GetWd()
	return newWatcher(filenames, wd)
}

func newWatcher(filenames []string, wd string) (*fileWatcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	watcher := &fileWatcher{
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

func (w *fileWatcher) addWatchRecursively(name string) error {
	if isDir, err := w.isDir(name); err == nil { // only adds if name represents a directory
		if !isDir {
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

func (w *fileWatcher) isDir(name string) (bool, error) {
	s, err := os.Stat(name)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, err
		}
	} else {
		return s.IsDir(), nil
	}
}

func (w *fileWatcher) addListener(listener fileChangeListener) {
	w.listeners = append(w.listeners, listener)
}

func (w *fileWatcher) watch() {
	defer func() {
		if err := w.watcher.Close(); err != nil {
			log.Println("[loader  ] fail to close watcher:", err)
		}
	}()

	atomic.StoreInt32(w.watching, 1)
	for atomic.LoadInt32(w.watching) == 1 { // w.watching is a flag indicates if keep watching or not
		if w.doWatch() {
			atomic.StoreInt32(w.watching, 0)
			break
		}
	}
}

func (w *fileWatcher) doWatch() (stop bool) {
	ok := true
	var event fsnotify.Event
	var err error

	select {
	case event, ok = <-w.watcher.Events:
		if ok {
			name := event.Name
			if !filepath.IsAbs(name) { // ensures absolute path
				if abs, err := filepath.Abs(name); err == nil {
					name = abs
				} else {
					log.Fatalln("[loader  ] fail to retrieve absolute path:", err)
				}
			}

			if event.Op&fsnotify.Create == fsnotify.Create {
				if err := w.addWatchRecursively(name); err != nil { // adds the newly-created file
					log.Fatalln("[loader  ] fail to add new file for automatic reloading:", err)
				}
			}
			if w.isConcernedFile(name) {
				w.notifyAll(name)
			}
		}
	case err, ok = <-w.watcher.Errors:
		if ok {
			log.Println("[loader  ] failure encountered when watching files:", err)
		}
	default:
		time.Sleep(watchInterval)
	}

	stop = !ok
	return
}

func (w *fileWatcher) isConcernedFile(name string) bool {
	for _, f := range w.filenames {
		af := f
		if !filepath.IsAbs(f) {
			af = filepath.Join(w.wd, f)
		}
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

func (w *fileWatcher) notifyAll(name string) {
	for _, l := range w.listeners {
		l.onFileChange(name)
	}
}

func (w *fileWatcher) cancel() {
	atomic.StoreInt32(w.watching, 0)
	time.Sleep(watchInterval)
}

func (l *Loader) EnableAutoReload(callback func(mappings *mckmaps.MockuMappings)) error {
	l.mux.Lock()
	defer l.mux.Unlock()

	if l.loaded == nil {
		panic("mockuMappings has not been loaded")
	}

	var err error

	if l.zipMode {
		wd := filepath.Dir(l.filename)
		l.watcher, err = newWatcher([]string{l.filename}, wd)
	} else {
		l.watcher, err = newWdWatcher(l.loaded.Filenames)
	}
	if err != nil {
		return err
	}

	listener := &autoReloadListener{
		l:        l,
		w:        l.watcher,
		callback: callback,
	}
	l.watcher.addListener(listener)
	go l.watcher.watch()

	return nil
}

type autoReloadListener struct {
	l         *Loader
	w         *fileWatcher
	callback  func(mappings *mckmaps.MockuMappings)
	reloadMux sync.Mutex
}

func (l *autoReloadListener) onFileChange(path string) {
	log.Println("[loader  ] changed  :", path)

	mappings, err := l.l.Load()
	if err != nil {
		log.Println("[loader  ] cannot load mockuMappings after changing:", err)
		return
	}

	// there can be only one goroutine reloading at the same time
	l.reloadMux.Lock()
	defer l.reloadMux.Unlock()

	// starts a new watcher goroutine, preventing from exiting
	if err := l.l.EnableAutoReload(l.callback); err != nil {
		log.Fatalln("[loader  ] fail to enable automatic reloading:", err)
	}
	go l.callback(mappings)

	l.w.cancel()
}
