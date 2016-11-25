package routes

type NotFoundError struct {
	error
}

type Blobstore interface {
	Get(path string) (body io.ReadCloser, redirectLocation string, err error)
	Put(path string, src io.ReadSeeker) (redirectLocation string, err error)
	Exists(path string) (bool, error)
}

type SignURLHandler interface {
	Sign(responseWriter http.ResponseWriter, request *http.Request)
}
