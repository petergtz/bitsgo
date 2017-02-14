package webdav

import (
	"crypto/tls"
	"io"
	"net/http"

	"github.com/petergtz/bitsgo/config"
	"github.com/petergtz/bitsgo/httputil"
	"github.com/petergtz/bitsgo/logger"
	"github.com/petergtz/bitsgo/routes"
	"github.com/pkg/errors"
	"github.com/uber-go/zap"
)

type Blobstore struct {
	httpClient     *http.Client
	webdavEndpoint string
	signer         *WebdavResourceSigner
}

func NewBlobstore(webdavEndpoint string) *Blobstore {
	return &Blobstore{
		webdavEndpoint: webdavEndpoint,
		// TODO: add correct timeouts etc. Don't skip certifcate verification!
		httpClient: &http.Client{
			Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
		},
		signer: NewWebdavResourceSigner(config.WebdavBlobstoreConfig{Endpoint: webdavEndpoint}),
	}
}

func (blobstore *Blobstore) Exists(path string) (bool, error) {
	url := blobstore.webdavEndpoint + "/" + path
	logger.Log.Debug("Exists", zap.String("path", path), zap.String("url", url))
	request, _ := http.NewRequest("HEAD", url, nil)
	request.SetBasicAuth("blobstore", "blobstore")
	response, e := blobstore.httpClient.Do(request)
	if e != nil {
		return false, errors.Wrapf(e, "Error in Exists, path=%v", path)
	}
	if response.StatusCode == http.StatusOK {
		logger.Log.Debug("Exists", zap.Bool("result", true))
		return true, nil
	}
	logger.Log.Debug("Exists", zap.Bool("result", false))
	return false, nil
}

func (blobstore *Blobstore) Get(path string) (body io.ReadCloser, redirectLocation string, err error) {
	exists, e := blobstore.Exists(path)
	if e != nil {
		return nil, "", e
	}
	if !exists {
		return nil, "", routes.NewNotFoundError()
	}
	signedUrl := blobstore.signer.Sign(path, "get")
	// signedUrl, e := blobstore.requestSignedWebdavUrl(path)
	if e != nil {
		return nil, "", e
	}
	return nil, signedUrl, nil
}

func (blobstore *Blobstore) Head(path string) (redirectLocation string, err error) {
	_, redirectLocation, e := blobstore.Get(path)
	return redirectLocation, e
}

func (blobstore *Blobstore) Put(path string, src io.ReadSeeker) (redirectLocation string, err error) {
	request, e := http.NewRequest("PUT", blobstore.webdavEndpoint+"/admin/"+path, src)
	if e != nil {
		return "", e
	}

	request.SetBasicAuth("blobstore", "blobstore")
	response, e := blobstore.httpClient.Do(request)
	if e != nil {
		return "", errors.Wrapf(e, "TODO")
	}
	// TODO improve error handling. It should provide a better error to the caller, so that the caller can provide
	//      a better response to its caller
	if response.StatusCode < 200 || response.StatusCode > 204 {
		return "", errors.Errorf("Expected StatusCreated, but got status code: " + response.Status)
	}
	return "", nil
}

func (blobstore *Blobstore) Copy(src, dest string) (redirectLocation string, err error) {
	logger.Log.Debug("Copy", zap.String("bla", blobstore.webdavEndpoint+"/admin/"+src))
	request, e := http.NewRequest("COPY", blobstore.webdavEndpoint+"/admin/"+src, nil)
	if e != nil {
		return "", errors.Wrapf(e, "TODO")
	}
	request.Header.Add("Destination", blobstore.webdavEndpoint+"/admin/"+dest)

	request.SetBasicAuth("blobstore", "blobstore")
	response, e := blobstore.httpClient.Do(request)
	if e != nil {
		return "", errors.Wrapf(e, "TODO")
	}
	// TODO improve error handling. It should provide a better error to the caller, so that the caller can provide
	//      a better response to its caller
	if response.StatusCode < 200 || response.StatusCode > 204 {
		return "", errors.Errorf("Expected StatusCreated, but got status code: " + response.Status)
	}
	return "", nil
}

func (blobstore *Blobstore) Delete(path string) error {
	request, e := http.NewRequest("DELETE", blobstore.webdavEndpoint+"/admin/"+path, nil)
	if e != nil {
		return errors.Wrapf(e, "TODO")
	}
	request.SetBasicAuth("blobstore", "blobstore")
	response, e := blobstore.httpClient.Do(request)
	if e != nil {
		return errors.Wrapf(e, "TODO")
	}
	// TODO improve error handling. It should provide a better error to the caller, so that the caller can provide
	//      a better response to its caller
	if response.StatusCode < 200 || response.StatusCode > 204 {
		return errors.Errorf("Expected StatusCreated, but got status code: " + response.Status)
	}
	return nil
}

func (blobstore *Blobstore) DeletePrefix(prefix string) error {
	if prefix != "" {
		prefix += "/"
	}
	request, e := http.NewRequest("DELETE", blobstore.webdavEndpoint+"/admin/"+prefix, nil)
	if e != nil {
		return errors.Wrapf(e, "TODO")
	}
	request.SetBasicAuth("blobstore", "blobstore")
	response, e := blobstore.httpClient.Do(request)
	if e != nil {
		return errors.Wrapf(e, "TODO")
	}

	if response.StatusCode == http.StatusNotFound {
		return routes.NewNotFoundError()
	}
	// TODO improve error handling. It should provide a better error to the caller, so that the caller can provide
	//      a better response to its caller
	if response.StatusCode < 200 || response.StatusCode > 204 {
		return errors.Errorf("Expected StatusCreated, but got status code: " + response.Status)
	}
	return nil
}

type NoRedirectBlobstore struct {
	httpClient     *http.Client
	webdavEndpoint string
	signer         *WebdavResourceSigner
}

func NewNoRedirectBlobstore(webdavEndpoint string) *NoRedirectBlobstore {
	return &NoRedirectBlobstore{
		webdavEndpoint: webdavEndpoint,
		// TODO: add correct timeouts etc. Don't skip certifcate verification!
		httpClient: &http.Client{
			Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
		},
		signer: NewWebdavResourceSigner(config.WebdavBlobstoreConfig{Endpoint: webdavEndpoint}),
	}
}

func (blobstore *NoRedirectBlobstore) Exists(path string) (bool, error) {
	url := blobstore.webdavEndpoint + "/" + path
	logger.Log.Debug("Exists", zap.String("path", path), zap.String("url", url))
	response, e := blobstore.httpClient.Head(url)
	if e != nil {
		return false, errors.Wrapf(e, "Error in Exists, path=%v", path)
	}
	if response.StatusCode == http.StatusOK {
		return true, nil
	}
	return false, nil
}

func (blobstore *NoRedirectBlobstore) Get(path string) (body io.ReadCloser, redirectLocation string, err error) {
	exists, e := blobstore.Exists(path)
	if e != nil {
		return nil, "", e
	}
	if !exists {
		return nil, "", routes.NewNotFoundError()
	}

	response, e := blobstore.httpClient.Get(blobstore.webdavEndpoint + "/" + path)

	if response.StatusCode != http.StatusOK {
		return nil, "", errors.Wrapf(e, "")
	}

	return response.Body, "", nil
}

// func (blobstore *Blobstore) requestSignedWebdavUrl(path string) (string, error) {
// 	blobstore.httpClient.Get(filepath.Join(blobstore.webdavEndpoint, "sign"))
// 	return "", nil
// }

func (blobstore *NoRedirectBlobstore) Head(path string) (redirectLocation string, err error) {
	_, redirectLocation, e := blobstore.Get(path)
	return redirectLocation, e
}

func (blobstore *NoRedirectBlobstore) Put(path string, src io.ReadSeeker) (redirectLocation string, err error) {
	request, e := httputil.NewPutRequest(blobstore.webdavEndpoint+"/admin/"+path, map[string]map[string]io.Reader{
		"dummy": map[string]io.Reader{"dummyfilename": src},
	})
	if e != nil {
		return "", e
	}

	request.SetBasicAuth("blobstore", "blobstore")
	response, e := blobstore.httpClient.Do(request)
	if e != nil {
		return "", errors.Wrapf(e, "TODO")
	}
	// TODO improve error handling. It should provide a better error to the caller, so that the caller can provide
	//      a better response to its caller
	if response.StatusCode < 200 || response.StatusCode > 204 {
		return "", errors.Errorf("Expected StatusCreated, but got status code: " + response.Status)
	}
	return "", nil
}

func (blobstore *NoRedirectBlobstore) Copy(src, dest string) (redirectLocation string, err error) {
	request, e := http.NewRequest("COPY", blobstore.webdavEndpoint+"/admin/"+src, nil)
	if e != nil {
		return "", errors.Wrapf(e, "TODO")
	}
	request.Header.Add("Destination", "/"+dest)

	request.SetBasicAuth("blobstore", "blobstore")
	response, e := blobstore.httpClient.Do(request)
	if e != nil {
		return "", errors.Wrapf(e, "TODO")
	}
	// TODO improve error handling. It should provide a better error to the caller, so that the caller can provide
	//      a better response to its caller
	if response.StatusCode < 200 || response.StatusCode > 204 {
		return "", errors.Errorf("Expected StatusCreated, but got status code: " + response.Status)
	}
	return "", nil
}

func (blobstore *NoRedirectBlobstore) Delete(path string) error {
	// request, e := http.NewRequest("COPY", blobstore.webdavEndpoint+"/admin/"+src, nil)
	// if e != nil {
	// 	return "", errors.Wrapf(e, "TODO")
	// }
	// request.Header.Add("Destination", "/"+dest)

	// request.SetBasicAuth("blobstore", "blobstore")
	// response, e := blobstore.httpClient.Do(request)
	// if e != nil {
	// 	return "", errors.Wrapf(e, "TODO")
	// }
	// // TODO improve error handling. It should provide a better error to the caller, so that the caller can provide
	// //      a better response to its caller
	// if response.StatusCode != http.StatusNoContent {
	// 	return "", errors.Errorf("Expected StatusCreated, but got status code: " + response.Status)
	// }
	return nil
}

func (blobstore *NoRedirectBlobstore) DeletePrefix(prefix string) error {
	return nil
}
