package loader

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/kumasuke120/mockuma/internal/mckmaps"
	"github.com/kumasuke120/mockuma/internal/myos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type lForTestFileChangeWatcher struct {
	okChan chan bool
	name   string
	bytes  []byte
}

func (l *lForTestFileChangeWatcher) onFileChange(path string) {
	fmt.Println("changed: " + path)

	if l.name != path {
		l.okChan <- false
		return
	}
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		l.okChan <- false
		return
	}
	if !reflect.DeepEqual(l.bytes, bytes) {
		l.okChan <- false
		return
	}
	l.okChan <- true
}

//noinspection GoImportUsedAsName
func TestFileWatcher(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	require.Nil(myos.InitWd())
	oldWd := myos.GetWd()

	dir, err := ioutil.TempDir("", "test-wd")
	require.Nil(err)
	require.Nil(myos.Chdir(dir))

	f1, e1 := ioutil.TempFile(dir, "fileWatcher")
	require.Nil(e1)
	n1 := f1.Name()
	fmt.Println(n1)

	require.Nil(f1.Close())

	expected := []byte{0xCA, 0xFE, 0xBA, 0xBE}

	assert.Panics(func() {
		_, _ = newWdWatcher(nil)
	})
	assert.Panics(func() {
		_, _ = newWdWatcher([]string{n1})
	})
	rn1, e1 := filepath.Rel(dir, n1)
	require.Nil(e1)

	time.Sleep(watchInterval * 2)
	w1, e1 := newWdWatcher([]string{rn1})
	require.Nil(e1)

	l := &lForTestFileChangeWatcher{
		okChan: make(chan bool, 10), // same event maybe triggered multiple times
		name:   n1,
		bytes:  expected,
	}
	w1.addListener(l)
	go w1.watch()

	assert.Nil(w1.addWatchRecursively(filepath.Join(dir, "not_exists")))

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			w := atomic.LoadInt32(w1.watching)
			if w == int32(1) {
				break
			}
			time.Sleep(watchInterval * 2)
		}
	}()

	time.Sleep(watchInterval * 2)
	wg.Wait()
	require.Nil(ioutil.WriteFile(n1, expected, 0644))
	assert.True(<-l.okChan)

	go func() {
		w1.watcher.Events <- fsnotify.Event{
			Name: rn1,
			Op:   fsnotify.Create,
		}
	}()
	assert.True(<-l.okChan)

	go func() {
		w1.watcher.Errors <- errors.New("test")
	}()
	time.Sleep(watchInterval * 2)

	w1.cancel()
	time.Sleep(watchInterval * 2)
	assert.Equal(int32(0), atomic.LoadInt32(w1.watching))

	require.Nil(os.Chdir(filepath.Join(oldWd, "testdata")))
	time.Sleep(watchInterval * 2)
	w2, e2 := newWdWatcher([]string{"mappings-0.json"})
	require.Nil(e2)
	go w2.watch()

	time.Sleep(watchInterval * 2)
	require.Nil(w2.watcher.Close())
	time.Sleep(watchInterval * 2)
	assert.Equal(int32(0), atomic.LoadInt32(w2.watching))

	require.Nil(myos.Chdir(oldWd))
	require.Nil(os.Remove(n1))
	require.Nil(os.RemoveAll(dir))
}

// test for normal loading
//noinspection GoImportUsedAsName
func TestLoader_EnableAutoReload(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	require.Nil(myos.InitWd())
	oldWd := myos.GetWd()

	dir, err := ioutil.TempDir("", "test-autoReload")
	require.Nil(err)
	f1, e1 := ioutil.TempFile(dir, "enableAutoReload")
	require.Nil(e1)
	n1 := f1.Name()
	fmt.Println(n1)

	_, e1 = f1.Write([]byte(`{"type":"main","include":{"mappings":[]}}`))
	require.Nil(e1)
	require.Nil(f1.Close())

	ld := New(n1)
	assert.Panics(func() {
		_ = ld.EnableAutoReload(nil)
	})
	_, e1 = ld.Load()
	assert.Nil(e1)

	rn1, e1 := filepath.Rel(dir, n1)
	require.Nil(e1)
	okChan := make(chan bool)
	e1 = ld.EnableAutoReload(func(m *mckmaps.MockuMappings) {
		okChan <- m != nil && m.Filenames[0] == rn1
	})
	assert.Nil(e1)

	time.Sleep(watchInterval * 2)
	require.Nil(ioutil.WriteFile(n1, []byte(`{"type": "main","include": {"mappings": []}}`), 0644))
	assert.True(<-okChan)

	time.Sleep(watchInterval * 2)
	require.Nil(ioutil.WriteFile(n1, []byte(`{}`), 0644))
	select {
	case _ = <-okChan:
		t.Fatal("'okChan' should be empty")
	default:
		t.Log("'okChan' is correct")
	}
	time.Sleep(watchInterval * 2)

	require.Nil(myos.Chdir(oldWd))
	require.Nil(os.Remove(n1))
	require.Nil(os.RemoveAll(dir))
}

// test for loading in zip mode
//noinspection GoImportUsedAsName
func TestLoader_EnableAutoReload2(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	require.Nil(myos.InitWd())

	fp := filepath.Join("testdata", "mappings.zip")
	abs, err := filepath.Abs(fp)
	require.Nil(err)

	ld := New(abs)
	_, err = ld.Load()
	if assert.Nil(err) {
		assert.True(ld.zipMode)
	}

	okChan := make(chan bool)
	err = ld.EnableAutoReload(func(m *mckmaps.MockuMappings) {
		okChan <- m != nil && assert.NotEmpty(m.Filenames)
	})
	assert.Nil(err)

	go func() {
		ld.watcher.watcher.Events <- fsnotify.Event{
			Name: abs,
			Op:   fsnotify.Write,
		}
	}()
	time.Sleep(watchInterval * 2)
	assert.True(<-okChan)
	err = ld.Clean()
	assert.Nil(err)
}
