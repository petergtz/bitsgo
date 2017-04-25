package ruby

import (
	"io"
	"os/exec"

	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/pkg/errors"
)

type Blobstore struct {
	fogConfig      string
	rubyScriptPath string
}

func NewBlobstore(fogConfig string, rubyScriptPath string) *Blobstore {
	return &Blobstore{fogConfig, rubyScriptPath}
}

func (blobstore *Blobstore) Exists(path string) (bool, error) {
	out, e := exec.Command("bundle", "exec", blobstore.rubyScriptPath+"/fogclient.rb", blobstore.fogConfig, "Exists", path).CombinedOutput()
	if e != nil {
		// fmt.Printf("%s", out)
		return false, errors.Wrapf(e, "Failed to execute Ruby subprocess. Path: %v. Stdout: %s", path, out)
	}
	switch string(out) {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, errors.Wrapf(e, "Received unexpected result ('%s') from Ruby for path '%v'", out, path)
	}
}

func (blobstore *Blobstore) HeadOrRedirectAsGet(path string) (redirectLocation string, err error) {
	// return signedURLFrom(request, blobstore.bucket, path)
	return
}

func (blobstore *Blobstore) Get(path string) (body io.ReadCloser, err error) {
	return
}

func (blobstore *Blobstore) GetOrRedirect(path string) (body io.ReadCloser, redirectLocation string, err error) {
	return
}

func (blobstore *Blobstore) Put(path string, src io.ReadSeeker) error {
	return nil
}

func (blobstore *Blobstore) PutOrRedirect(path string, src io.ReadSeeker) (redirectLocation string, err error) {
	// This is the behavior as in the current Ruby implementation
	e := blobstore.Put(path, src)
	return "", e
}

func signedURLFrom(req *request.Request, bucket, path string) (string, error) {
	// return signedURL, nil
	return "", nil
}

func (blobstore *Blobstore) Copy(src, dest string) error {
	return nil
}

func (blobstore *Blobstore) Delete(path string) error {
	return nil
}

func (blobstore *Blobstore) DeleteDir(prefix string) error {
	return nil
}
