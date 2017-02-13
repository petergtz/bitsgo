package webdav

import (
	"crypto/tls"
	"net/http"
	"strings"

	"fmt"
	"time"

	"io/ioutil"

	"github.com/petergtz/bitsgo/config"
	"github.com/petergtz/bitsgo/httputil"
)

func NewWebdavResourceSigner(config config.WebdavBlobstoreConfig) *WebdavResourceSigner {
	return &WebdavResourceSigner{
		webdavEndpoint: config.Endpoint,
		// TODO: add correct timeouts etc.
		httpClient: &http.Client{
			// TODO skip SSL validation
			Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
		},
	}
}

type WebdavResourceSigner struct {
	httpClient     *http.Client
	webdavEndpoint string
}

func (signer *WebdavResourceSigner) Sign(resource string, method string) string {
	var url string
	switch strings.ToLower(method) {
	case "put":
		// TODO why do we need a "/" before the resource?
		url = fmt.Sprintf(signer.webdavEndpoint+"/sign_for_put?path=/%v&expires=%v", resource, time.Now().Unix()+3600)
	case "get":
		url = fmt.Sprintf(signer.webdavEndpoint+"/sign?path=/%v&expires=%v", resource, time.Now().Unix()+3600)
	}
	request, _ := http.NewRequest("GET", url, nil)
	request.SetBasicAuth("blobstore", "blobstore")
	response, e := signer.httpClient.Do(request)
	if e != nil {
		return "Error during signing. Error: " + e.Error()
	}
	if response.StatusCode != http.StatusOK {
		return "Error during signing. Error code: " + response.Status
	}
	defer response.Body.Close()
	content, e := ioutil.ReadAll(response.Body)
	if e != nil {
		return "Error reading response body. Error: " + e.Error()
	}

	signedUrl := httputil.MustParse(string(content))

	// TODO Is this really what we want to do?
	signedUrl.Host = httputil.MustParse(signer.webdavEndpoint).Host
	signedUrl.Scheme = httputil.MustParse(signer.webdavEndpoint).Scheme

	return signedUrl.String()
}
