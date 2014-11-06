package main

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"time"
)

func createSpace(mRetrieveConfPath string, done chan<- bool) (handler http.HandlerFunc, err error) {
	if mRetrieveConfPath != "" {
		f, err := os.Open(mRetrieveConfPath)
		if err != nil {
			return handler, err
		}

		handler = func(w http.ResponseWriter, r *http.Request) {
			defer f.Close()
			http.ServeContent(w, r, "mms", time.Time{}, f)
			done <- true
		}
	} else {
		handler = func(w http.ResponseWriter, r *http.Request) {
			http.ServeContent(w, r, "mms", time.Time{}, bytes.NewReader(mRetrieveConf))
			done <- true
		}
	}

	return handler, err
}

func copyMMS(mRetrieveConfPath, mmsPath string) error {
	src, err := os.Open(mRetrieveConfPath)
	if err != nil {
		return err
	}

	dst, err := os.Create(mmsPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}
