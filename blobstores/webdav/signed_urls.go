package webdav

import (
	"net/http"

	"github.com/petergtz/bitsgo/config"
)

func NewWebdavResourceSigner(config config.WebdavBlobstoreConfig) *WebdavResourceSigner {
	return &WebdavResourceSigner{
		webdavEndpoint: config.Endpoint,
		// TODO: add correct timeouts etc.
		httpClient: &http.Client{},
	}
}

type WebdavResourceSigner struct {
	httpClient     *http.Client
	webdavEndpoint string
}

func (signer *WebdavResourceSigner) Sign(resource string, method string) (signedURL string) {
	return "TODO: signedURL"
}
