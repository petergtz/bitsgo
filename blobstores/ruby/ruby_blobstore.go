package ruby

import (
	"fmt"
	"io"
	"os/exec"

	"bytes"

	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/pkg/errors"
)

type Blobstore struct {
	fogConfig      string
	rubyScriptPath string
	directoryKey   string
}

func NewBlobstore(fogConfig string, rubyScriptPath string, directoryKey string) *Blobstore {
	return &Blobstore{fogConfig, rubyScriptPath, directoryKey}
}

func (blobstore *Blobstore) Exists(path string) (bool, error) {
	command := exec.Command("bundle", "exec", blobstore.rubyScriptPath+"/fogclient.rb", blobstore.fogConfig, "Exists", blobstore.directoryKey, path)
	stderr := &bytes.Buffer{}
	command.Stderr = stderr
	out, e := command.Output()
	if e != nil {
		return false, errors.Wrapf(e, "Failed to execute Ruby subprocess. Path: %v. Stdout: %s. Stderr: %v", path, out, stderr.String())
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

// func (blobstore *Blobstore) runRuby(args ...string) {
// 	command := exec.Command("bundle", "exec", blobstore.rubyScriptPath+"/fogclient.rb", blobstore.fogConfig, args...)
// 	stderr := &bytes.Buffer{}
// 	command.Stderr = stderr
// 	out, e := command.Output()
// 	if e != nil {
// 		return false, errors.Wrapf(e, "Failed to execute Ruby subprocess. Path: %v. Stdout: %s. Stderr: %v", path, out, stderr.String())
// 	}
// }

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
	tempFile, e := ioutil.TempFile("", "")
	if e != nil {
		panic("TODO")
	}
	defer tempFile.Close()
	//  defer remove file
	io.Copy(tempFile, src) // TODO error handling
	command := exec.Command("bundle", "exec", blobstore.rubyScriptPath+"/fogclient.rb", blobstore.fogConfig, "Put", blobstore.directoryKey, path, tempFile.Name())
	fmt.Println(command.Args)
	// logger.Log.Debug("", zap.String
	stderr := &bytes.Buffer{}
	command.Stderr = stderr
	out, e := command.Output()
	if e != nil {
		return errors.Wrapf(e, "Failed to execute Ruby subprocess. Path: %v. Stdout: %s. Stderr: %v", path, out, stderr.String())
	}
	fmt.Println("out", string(out))
	if string(out) != "" {
		fmt.Println("XXXXX")
		return errors.Wrapf(e, "Received unexpected result ('%s') from Ruby for path '%v'", out, path)
	}
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
