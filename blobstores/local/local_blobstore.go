package local

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/petergtz/bitsgo/logger"
	"github.com/petergtz/bitsgo/routes"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/uber-go/zap"
)

type Blobstore struct {
	pathPrefix string
	fs         afero.Fs
}

func NewBlobstore(pathPrefix string) *Blobstore {
	return &Blobstore{
		pathPrefix: pathPrefix,
		fs:         afero.NewOsFs(),
	}
}

func NewBlobstoreWithFs(fs afero.Fs) *Blobstore {
	return &Blobstore{
		pathPrefix: "/",
		fs:         fs,
	}
}

func (blobstore *Blobstore) Exists(path string) (bool, error) {
	_, err := blobstore.fs.Stat(filepath.Join(blobstore.pathPrefix, path))
	blobstore.fs.
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("Could not stat on %v. Caused by: %v", filepath.Join(blobstore.pathPrefix, path), err)
	}
	return true, nil
}

func (blobstore *Blobstore) Get(path string) (body io.ReadCloser, redirectLocation string, err error) {
	logger.Log.Debug("Get", zap.String("local-path", filepath.Join(blobstore.pathPrefix, path)))
	file, e := os.Open(filepath.Join(blobstore.pathPrefix, path))

	if os.IsNotExist(e) {
		return nil, "", routes.NewNotFoundError()
	}
	if e != nil {
		return nil, "", fmt.Errorf("Error while opening file %v. Caused by: %v", path, e)
	}
	return file, "", nil
}

func (blobstore *Blobstore) Head(path string) (redirectLocation string, err error) {
	logger.Log.Debug("Head", zap.String("local-path", filepath.Join(blobstore.pathPrefix, path)))
	_, e := os.Stat(filepath.Join(blobstore.pathPrefix, path))

	if os.IsNotExist(e) {
		return "", routes.NewNotFoundError()
	}
	if e != nil {
		return "", fmt.Errorf("Error while opening file %v. Caused by: %v", path, e)
	}
	return "", nil
}

func (blobstore *Blobstore) Put(path string, src io.ReadSeeker) (redirectLocation string, err error) {
	e := os.MkdirAll(filepath.Dir(filepath.Join(blobstore.pathPrefix, path)), os.ModeDir|0755)
	if e != nil {
		return "", fmt.Errorf("Error while creating directories for %v. Caused by: %v", path, e)
	}
	file, e := os.Create(filepath.Join(blobstore.pathPrefix, path))
	if e != nil {
		return "", fmt.Errorf("Error while creating file %v. Caused by: %v", path, e)
	}
	defer file.Close()
	_, e = io.Copy(file, src)
	if e != nil {
		return "", fmt.Errorf("Error while writing file %v. Caused by: %v", path, e)
	}
	return "", nil
}

func (blobstore *Blobstore) Copy(src, dest string) (redirectLocation string, err error) {
	srcFull := filepath.Join(blobstore.pathPrefix, src)
	destFull := filepath.Join(blobstore.pathPrefix, dest)

	srcFile, e := os.Open(srcFull)
	if e != nil {
		if os.IsNotExist(e) {
			return "", routes.NewNotFoundError()
		}
		return "", errors.Wrapf(e, "Opening src failed. (src=%v, dest=%v)", srcFull, destFull)
	}
	defer srcFile.Close()

	e = os.MkdirAll(filepath.Dir(destFull), 0755)
	if e != nil {
		return "", errors.Wrapf(e, "Make dir failed. (src=%v, dest=%v)", srcFull, destFull)
	}

	destFile, e := os.Create(destFull)
	if e != nil {
		return "", errors.Wrapf(e, "Creating dest failed. (src=%v, dest=%v)", srcFull, destFull)
	}
	defer destFile.Close()

	_, e = io.Copy(destFile, srcFile)
	if e != nil {
		return "", errors.Wrapf(e, "Copying failed. (src=%v, dest=%v)", srcFull, destFull)
	}

	return "", nil
}

func (blobstore *Blobstore) Delete(path string) error {
	_, e := os.Stat(filepath.Join(blobstore.pathPrefix, path))
	if os.IsNotExist(e) {
		return routes.NewNotFoundError()
	}
	e = os.RemoveAll(filepath.Join(blobstore.pathPrefix, path))
	if e != nil {
		return fmt.Errorf("Error while deleting file %v. Caused by: %v", path, e)
	}
	return nil
}

func (blobstore *Blobstore) DeletePrefix(prefix string) error {
	// TODO this not strictly deleting a prefix. It assumes the prefix to be a directory.
	e := os.RemoveAll(filepath.Join(blobstore.pathPrefix, prefix))
	if e != nil {
		return errors.Wrapf(e, "Failed to delete path %v", filepath.Join(blobstore.pathPrefix, prefix))
	}
	return nil
}
