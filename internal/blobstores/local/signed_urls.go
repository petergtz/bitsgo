package local

import (
	"fmt"
	"net/http"

	"time"

	"github.com/petergtz/bitsgo/internal/pathsigner"
)

type SignatureVerificationMiddleware struct {
	SignatureValidator pathsigner.PathSignatureValidator
}

func (middleware *SignatureVerificationMiddleware) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request, next http.HandlerFunc) {
	if request.URL.Query().Get("md5") == "" {
		responseWriter.WriteHeader(403)
		return
	}
	if !middleware.SignatureValidator.SignatureValid(request.URL) {
		responseWriter.WriteHeader(403)
		return
	}
	next(responseWriter, request)
}

type LocalResourceSigner struct {
	Signer             pathsigner.PathSigner
	ResourcePathPrefix string
	DelegateEndpoint   string
}

func (signer *LocalResourceSigner) Sign(resource string, method string, expirationTime time.Time) (signedURL string) {
	return fmt.Sprintf("%s%s", signer.DelegateEndpoint, signer.Signer.Sign(signer.ResourcePathPrefix+resource, expirationTime))
}
