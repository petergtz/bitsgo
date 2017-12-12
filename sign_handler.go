package bitsgo

import (
	"fmt"
	"net/http"
	"time"

	"github.com/benbjohnson/clock"
)

type ResourceSigner interface {
	Sign(resource string, method string, expirationTime time.Time) (signedURL string)
}

type SignResourceHandler struct {
	clock                                clock.Clock
	putResourceSigner, getResourceSigner ResourceSigner
}

func NewSignResourceHandler(getResourceSigner, putResourceSigner ResourceSigner) *SignResourceHandler {
	return &SignResourceHandler{
		getResourceSigner: getResourceSigner,
		putResourceSigner: putResourceSigner,
		clock:             clock.New(),
	}
}

func (handler *SignResourceHandler) Sign(responseWriter http.ResponseWriter, request *http.Request, params map[string]string) {
	method := params["verb"]
	var signer ResourceSigner

	switch method {
	case "", "get":
		signer = handler.getResourceSigner
	case "put":
		signer = handler.putResourceSigner
	default:
		panic("Invalid method:" + method)
	}

	signature := signer.Sign(params["resource"], method, handler.clock.Now().Add(1*time.Hour))
	fmt.Fprint(responseWriter, signature)
}
