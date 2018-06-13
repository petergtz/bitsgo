package inmemory_blobstore

import (
	"fmt"
	"io"
	"strings"

	"github.com/petergtz/bitsgo/internal"

	"bytes"

	"io/ioutil"
)

type Blobstore struct {
	Entries map[string][]byte
}

func NewBlobstore() *Blobstore {
	return &Blobstore{Entries: make(map[string][]byte)}
}

func NewBlobstoreWithEntries(entries map[string][]byte) *Blobstore {
	return &Blobstore{Entries: entries}
}

func (blobstore *Blobstore) Exists(path string) (bool, error) {
	_, hasKey := blobstore.Entries[path]
	return hasKey, nil
}

func (blobstore *Blobstore) HeadOrRedirectAsGet(path string) (redirectLocation string, err error) {
	_, hasKey := blobstore.Entries[path]
	if !hasKey {
		return "", bitsgo.NewNotFoundError()
	}
	return "", nil
}

func (blobstore *Blobstore) Get(path string) (body io.ReadCloser, err error) {
	entry, hasKey := blobstore.Entries[path]
	if !hasKey {
		return nil, bitsgo.NewNotFoundError()
	}
	return ioutil.NopCloser(bytes.NewBuffer(entry)), nil
}

func (blobstore *Blobstore) GetOrRedirect(path string) (body io.ReadCloser, redirectLocation string, err error) {
	body, e := blobstore.Get(path)
	return body, "", e
}

func (blobstore *Blobstore) Put(path string, src io.ReadSeeker) error {
	b, e := ioutil.ReadAll(src)
	if e != nil {
		return fmt.Errorf("Error while reading from src %v. Caused by: %v", path, e)
	}
	blobstore.Entries[path] = b
	return nil
}

func (blobstore *Blobstore) Copy(src, dest string) error {
	blobstore.Entries[dest] = blobstore.Entries[src]
	return nil
}

func (blobstore *Blobstore) Delete(path string) error {
	_, hasKey := blobstore.Entries[path]
	if !hasKey {
		return bitsgo.NewNotFoundError()
	}
	delete(blobstore.Entries, path)
	return nil
}

func (blobstore *Blobstore) DeleteDir(prefix string) error {
	for key := range blobstore.Entries {
		if strings.HasPrefix(key, prefix) {
			delete(blobstore.Entries, key)
		}

	}
	return nil
}
