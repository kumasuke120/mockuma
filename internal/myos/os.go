package myos

import (
	"os"
	"sync"
)

var theWd = ""
var wdMutex sync.Mutex

func InitWd() error {
	wdMutex.Lock()
	defer wdMutex.Unlock()

	if len(theWd) == 0 {
		// set current working directory
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		theWd = wd
	}

	return nil
}

func setWd(wd string) {
	wdMutex.Lock()
	defer wdMutex.Unlock()

	theWd = wd
}

func GetWd() string {
	wdMutex.Lock()
	defer wdMutex.Unlock()

	return theWd
}

func Chdir(dir string) (err error) {
	if dir == GetWd() {
		return
	}

	err = os.Chdir(dir)
	setWd(dir)
	return
}
