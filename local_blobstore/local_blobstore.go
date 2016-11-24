package local_blobstore

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type LocalBlobstore struct {
	pathPrefix string
}

func NewLocalBlobstore(pathPrefix string) *LocalBlobstore {
	return &LocalBlobstore{pathPrefix: pathPrefix}
}

func (blobstore *LocalBlobstore) Exists(path string) (statusCode int) {
	_, err := os.Stat(filepath.Join(blobstore.pathPrefix, path))
	if os.IsNotExist(err) {
		return http.StatusNotFound
	}
	if err != nil {
		log.Printf("Could not stat on %v. Caused by: %v", filepath.Join(blobstore.pathPrefix, path), err)
		return http.StatusInternalServerError
	}
	return http.StatusOK
}
func (blobstore *LocalBlobstore) Get(path string) (statusCode int, body io.ReadCloser, header map[string][]string) {
	file, e := os.Open(filepath.Join(blobstore.pathPrefix, path))

	if os.IsNotExist(e) {
		return http.StatusNotFound, nil, make(map[string][]string)
	}
	if e != nil {
		log.Printf("Error while opening file %v. Caused by: %v", path, e)
		return http.StatusInternalServerError, nil, make(map[string][]string)
	}
	return http.StatusOK, file, make(map[string][]string)
}

func (blobstore *LocalBlobstore) Put(path string, src io.ReadSeeker) (statusCode int, header map[string][]string) {
	e := os.MkdirAll(filepath.Dir(filepath.Join(blobstore.pathPrefix, path)), os.ModeDir|0755)
	if e != nil {
		log.Printf("Error while creating directories for %v. Caused by: %v", path, e)
		return http.StatusInternalServerError, make(map[string][]string)
	}
	file, e := os.Create(filepath.Join(blobstore.pathPrefix, path))
	defer file.Close()
	if e != nil {
		log.Printf("Error while creating file %v. Caused by: %v", path, e)
		return http.StatusInternalServerError, make(map[string][]string)
	}
	_, e = io.Copy(file, src)
	if e != nil {
		log.Printf("Error while writing file %v. Caused by: %v", path, e)
		return http.StatusInternalServerError, make(map[string][]string)
	}
	return http.StatusCreated, make(map[string][]string)
}
