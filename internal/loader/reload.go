package loader

import (
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/kumasuke120/mockuma/internal/mckmaps"
)

type fileChangeListener interface {
	onFileChange(path string)
}

type fileChangeWatcher struct {
	watcher   *fsnotify.Watcher
	listeners []fileChangeListener
	watching  *int32
}

func newWatcher(filenames []string) (*fileChangeWatcher, error) {
	if len(filenames) == 0 {
		panic("parameter 'filenames' should not be empty")
	}

	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	// adds filenames to watch
	for _, f := range filenames {
		err := fsWatcher.Add(f)
		if err != nil {
			return nil, err
		}
	}

	watcher := &fileChangeWatcher{
		watcher:  fsWatcher,
		watching: new(int32),
	}
	return watcher, nil
}

func (w *fileChangeWatcher) addListener(listener fileChangeListener) {
	w.listeners = append(w.listeners, listener)
}

func (w *fileChangeWatcher) watch() {
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

			if event.Op&fsnotify.Write == fsnotify.Write {
				w.notifyAll(event.Name)
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

func (w *fileChangeWatcher) notifyAll(name string) {
	for _, l := range w.listeners {
		l.onFileChange(name)
	}
}

func (w *fileChangeWatcher) cancel() {
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
	w         *fileChangeWatcher
	callback  func(mappings *mckmaps.MockuMappings)
	reloadMux sync.Mutex
}

func (l *autoReloadListener) onFileChange(path string) {
	log.Println("[loader] change detected:", path)

	mappings, err := l.l.Load()
	if err != nil {
		log.Println("[loader] cannot load mockuMappings after changing:", err)
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
