package loader

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
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
func TestFileChangeWatcher(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	f1, e1 := ioutil.TempFile("", "wdWatcher")
	require.Nil(e1)
	n1 := f1.Name()
	fmt.Println(n1)

	require.Nil(f1.Close())

	expected := []byte{0xCA, 0xFE, 0xBA, 0xBE}
	w1, e1 := newWatcher([]string{n1})
	require.Nil(e1)

	l := &lForTestFileChangeWatcher{
		okChan: make(chan bool),
		name:   n1,
		bytes:  expected,
	}
	w1.addListener(l)
	go w1.watch()

	time.Sleep(1 * time.Second)
	require.Nil(ioutil.WriteFile(n1, expected, 0644))

	time.Sleep(1 * time.Second)
	assert.True(<-l.okChan)

	w1.cancel()
	time.Sleep(1 * time.Second)

	require.Nil(os.Remove(n1))
}

//noinspection GoImportUsedAsName
func TestLoader_EnableAutoReload(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	oldWd, err := os.Getwd()
	require.Nil(err)

	f1, e1 := ioutil.TempFile("", "enableAutoReload")
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

	okChan := make(chan bool)
	e1 = ld.EnableAutoReload(func(m *mckmaps.MockuMappings) {
		okChan <- m != nil && m.Filenames[0] == n1
	})
	assert.Nil(e1)

	time.Sleep(1 * time.Second)
	require.Nil(ioutil.WriteFile(n1, []byte(`{"@type": "main","@include": {"mappings": []}}`), 0644))

	assert.True(<-okChan)

	require.Nil(os.Chdir(oldWd))
	require.Nil(os.Remove(n1))
}
