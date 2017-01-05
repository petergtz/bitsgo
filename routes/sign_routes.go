package routes

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/petergtz/bitsgo/basic_auth_middleware"
	"github.com/urfave/negroni"
)

type SignURLHandler interface {
	Sign(responseWriter http.ResponseWriter, request *http.Request)
}

func SetUpSignRoute(router *mux.Router, basicAuthMiddleware *basic_auth_middleware.BasicAuthMiddleware,
	signPackageURLHandler, signDropletURLHandler, signBuildpackURLHandler SignURLHandler) {
	router.Path("/sign/packages/{guid}").Methods("GET").Handler(wrapWith(basicAuthMiddleware, signPackageURLHandler))
	router.Path("/sign/droplets/{guid:.*}").Methods("GET").Handler(wrapWith(basicAuthMiddleware, signDropletURLHandler))
	router.Path("/sign/buildpacks/{guid}").Methods("GET").Handler(wrapWith(basicAuthMiddleware, signBuildpackURLHandler))
	// TODO should this rather get its own handler instead of using the droplets' one?
	router.Path("/sign/buildpack_cache/entries/{guid:.*}").Methods("GET").Handler(wrapWith(basicAuthMiddleware, signDropletURLHandler))
}

func wrapWith(basicAuthMiddleware *basic_auth_middleware.BasicAuthMiddleware, handler SignURLHandler) http.Handler {
	return negroni.New(
		basicAuthMiddleware,
		negroni.Wrap(http.HandlerFunc(handler.Sign)),
	)
}
