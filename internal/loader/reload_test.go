package loader

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"github.com/kumasuke120/mockuma/internal/mckmaps"
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
func TestWdWatcher(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	oldWd, err := os.Getwd()
	require.Nil(err)

	dir := os.TempDir()
	require.Nil(os.Chdir(dir))

	f1, e1 := ioutil.TempFile(dir, "wdWatcher")
	require.Nil(e1)
	n1 := f1.Name()
	fmt.Println(n1)

	require.Nil(f1.Close())

	expected := []byte{0xCA, 0xFE, 0xBA, 0xBE}

	assert.Panics(func() {
		_, _ = newWatcher(nil)
	})
	assert.Panics(func() {
		_, _ = newWatcher([]string{n1})
	})
	rn1, e1 := filepath.Rel(dir, n1)
	require.Nil(e1)

	time.Sleep(1 * time.Second)
	w1, e1 := newWatcher([]string{rn1})
	require.Nil(e1)

	l := &lForTestFileChangeWatcher{
		okChan: make(chan bool, 10), // same event maybe triggered multiple times
		name:   n1,
		bytes:  expected,
	}
	w1.addListener(l)
	go w1.watch()

	time.Sleep(1 * time.Second)
	require.Nil(ioutil.WriteFile(n1, expected, 0644))

	time.Sleep(1 * time.Second)
	assert.True(<-l.okChan)

	go func() {
		w1.watcher.Errors <- errors.New("test")
	}()
	time.Sleep(1 * time.Second)

	w1.cancel()
	time.Sleep(1 * time.Second)
	assert.Equal(int32(0), atomic.LoadInt32(w1.watching))

	require.Nil(os.Chdir(filepath.Join(oldWd, "testdata")))
	time.Sleep(1 * time.Second)
	w2, e2 := newWatcher([]string{"mappings-0.json"})
	require.Nil(e2)
	go w2.watch()

	time.Sleep(1 * time.Second)
	require.Nil(w2.watcher.Close())
	time.Sleep(1 * time.Second)
	assert.Equal(int32(0), atomic.LoadInt32(w2.watching))

	require.Nil(os.Chdir(oldWd))
	require.Nil(os.Remove(n1))
}

//noinspection GoImportUsedAsName
func TestLoader_EnableAutoReload(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	oldWd, err := os.Getwd()
	require.Nil(err)

	dir := os.TempDir()
	f1, e1 := ioutil.TempFile(dir, "enableAutoReload")
	require.Nil(e1)
	n1 := f1.Name()
	fmt.Println(n1)

	_, e1 = f1.Write([]byte(`{"@type":"main","@include":{"mappings":[]}}`))
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

	time.Sleep(1 * time.Second)
	require.Nil(ioutil.WriteFile(n1, []byte(`{"@type": "main","@include": {"mappings": []}}`), 0644))
	assert.True(<-okChan)

	time.Sleep(1 * time.Second)
	require.Nil(ioutil.WriteFile(n1, []byte(`{}`), 0644))
	select {
	case _ = <-okChan:
		t.Fatal("'okChan' should be empty")
	default:
		t.Log("'okChan' is correct")
	}
	time.Sleep(1 * time.Second)

	require.Nil(os.Chdir(oldWd))
	require.Nil(os.Remove(n1))
}
