package loader

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func unzip(filename string) (string, error) {
	r, err := zip.OpenReader(filename)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := r.Close(); err != nil {
			log.Println("[loader  ] fail to close zip reader")
		}
	}()

	dir, err := createTempDir()
	if err != nil {
		return "", err
	}

	for _, f := range r.File {
		err := unzipToDir(f, dir)
		if err != nil {
			return "", err
		}
	}
	log.Println("[loader  ] unzip    : archive extracted to:", dir)

	return dir, nil
}

func createTempDir() (string, error) {
	dirPat := fmt.Sprintf("mockuma_%d_*", os.Getpid())
	return ioutil.TempDir(os.TempDir(), dirPat)
}

func unzipToDir(f *zip.File, dir string) error {
	srcFR, err := f.Open()
	if err != nil {
		return err
	}
	defer func() {
		if err := srcFR.Close(); err != nil {
			log.Println("[loader  ] fail to close a file in the zip reader")
		}
	}()

	dst := filepath.Join(dir, f.Name)
	if f.FileInfo().IsDir() {
		err := os.MkdirAll(dst, f.Mode())
		if err != nil {
			return err
		}
	} else {
		err := os.MkdirAll(filepath.Dir(dst), f.Mode())
		if err != nil {
			return err
		}
		dstF, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		defer func() {
			if err := dstF.Close(); err != nil {
				log.Fatalln("[loader  ] cannot close the newly-created file in temporary directory")
			}
		}()

		_, err = io.Copy(dstF, srcFR)
		if err != nil {
			return err
		}
	}
	return nil
}
