package main

import "net/http"

type SpaFileSystem struct {
	underlying http.FileSystem
}

func SpaFileServer(underlying http.FileSystem) *SpaFileSystem {
	return &SpaFileSystem{underlying}
}

func (fs *SpaFileSystem) Open(name string) (http.File, error) {
	f, err := fs.underlying.Open(name)

	if err != nil {
		f, err = fs.underlying.Open("/index.html")

		if err != nil {
			return nil, err
		}

		return f, nil
	}

	return f, nil
}
