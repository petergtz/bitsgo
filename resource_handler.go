package bitsgo

import (
	"bytes"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/petergtz/bitsgo/logger"
	"github.com/pkg/errors"
)

type MetricsService interface {
	SendTimingMetric(name string, duration time.Duration)
}

type StateForbiddenError struct {
	error
}

func NewStateForbiddenError() *StateForbiddenError {
	return &StateForbiddenError{fmt.Errorf("StateForbiddenError")}
}

type Updater interface {
	NotifyProcessingUpload(guid string) error
	NotifyUploadSucceeded(guid string, sha1 string, sha2 string) error
	NotifyUploadFailed(guid string, e error) error
}

type NullUpdater struct{}

func (u *NullUpdater) NotifyProcessingUpload(guid string) error                          { return nil }
func (u *NullUpdater) NotifyUploadSucceeded(guid string, sha1 string, sha2 string) error { return nil }
func (u *NullUpdater) NotifyUploadFailed(guid string, e error) error                     { return nil }

type ResourceHandler struct {
	blobstore        Blobstore
	resourceType     string
	metricsService   MetricsService
	maxBodySizeLimit uint64
	updater          Updater
}

type responseBody struct {
	Guid      string    `json:"guid"`
	State     string    `json:"state"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
}

func NewResourceHandler(blobstore Blobstore, resourceType string, metricsService MetricsService, maxBodySizeLimit uint64) *ResourceHandler {
	return &ResourceHandler{
		blobstore:        blobstore,
		resourceType:     resourceType,
		metricsService:   metricsService,
		maxBodySizeLimit: maxBodySizeLimit,
		updater:          &NullUpdater{},
	}
}

func NewResourceHandlerWithUpdater(blobstore Blobstore, updater Updater, resourceType string, metricsService MetricsService, maxBodySizeLimit uint64) *ResourceHandler {
	return &ResourceHandler{
		blobstore:        blobstore,
		resourceType:     resourceType,
		metricsService:   metricsService,
		maxBodySizeLimit: maxBodySizeLimit,
		updater:          updater,
	}
}

// TODO: instead of params, we could use `identifier string` to make the interface more type-safe.
//       Here and in the other methods.
func (handler *ResourceHandler) AddOrReplaceWithDigestInHeader(responseWriter http.ResponseWriter, request *http.Request, params map[string]string) {
	if !HandleBodySizeLimits(responseWriter, request, handler.maxBodySizeLimit) {
		return
	}
	logger.From(request).Debugw("Octet-Stream")

	digest := request.Header.Get("Digest")
	if digest == "" {
		badRequest(responseWriter, request, "No Digest header")
		return
	}
	parts := strings.Split(digest, "=")
	if len(parts) != 2 {
		badRequest(responseWriter, request, "Digest must have format sha256=value, but is "+digest)
		return
	}
	alg, value := strings.ToLower(parts[0]), parts[1]
	if alg != "sha256" {
		badRequest(responseWriter, request, "Digest must have format sha256=value, but is "+digest)
		return
	}
	if value == "" {
		badRequest(responseWriter, request, "Digest must have format sha256=value. Value cannot be empty")
		return
	}

	// TODO this can cause an out of memory panic. Should be smart about writing big files to disk instead.
	content, e := ioutil.ReadAll(request.Body)
	if e != nil {
		internalServerError(responseWriter, request, e)
		return
	}
	e = handler.blobstore.Put(params["identifier"]+"/"+value, bytes.NewReader(content))
	// TODO use Clock instead:
	writeResponseBasedOn("", e, responseWriter, request, http.StatusCreated, nil, &responseBody{Guid: params["identifier"], State: "READY", Type: "bits", CreatedAt: time.Now()}, "")
}

// TODO: instead of params, we could use `identifier string` to make the interface more type-safe.
//       Here and in the other methods.
func (handler *ResourceHandler) AddOrReplace(responseWriter http.ResponseWriter, request *http.Request, params map[string]string) {
	if !HandleBodySizeLimits(responseWriter, request, handler.maxBodySizeLimit) {
		return
	}
	file, _, e := request.FormFile(handler.resourceType)
	if e == http.ErrMissingFile {
		file, _, e = request.FormFile("bits")
	}
	if e == http.ErrMissingFile {
		badRequest(responseWriter, request, "Could not retrieve form parameter '%s' or 'bits", handler.resourceType)
		return
	}
	if e != nil {
		internalServerError(responseWriter, request, e)
		return
	}
	defer file.Close()

	tempFilename, e := CreateTempFileWithContent(file)
	if e != nil {
		internalServerError(responseWriter, request, e)
		return
	}

	e = handler.updater.NotifyProcessingUpload(params["identifier"])
	if handleNotificationError(e, responseWriter, request) {
		return
	}

	if request.URL.Query().Get("async") == "true" {
		go handler.uploadResource(tempFilename, request, params["identifier"], true)
		e = nil
	} else {
		e = handler.uploadResource(tempFilename, request, params["identifier"], false)
	}
	writeResponseBasedOn("", e, responseWriter, request, http.StatusCreated, nil, &responseBody{Guid: params["identifier"], State: "READY", Type: "bits", CreatedAt: time.Now()}, "")
}

func CreateTempFileWithContent(reader io.Reader) (string, error) {
	uploadedFile, e := ioutil.TempFile("", "bits")
	if e != nil {
		return "", errors.WithStack(e)
	}
	_, e = io.Copy(uploadedFile, reader)
	if e != nil {
		uploadedFile.Close()
		return "", errors.WithStack(e)
	}
	uploadedFile.Close()

	return uploadedFile.Name(), nil
}

func (handler *ResourceHandler) uploadResource(tempFilename string, request *http.Request, identifier string, async bool) error {
	defer os.Remove(tempFilename)

	tempFile, e := os.Open(tempFilename)
	if e != nil {
		handler.notifyUploadFailed(identifier, e, request)
		return handle(errors.Wrapf(e, "Could not open temporary file '%v'", tempFilename), async, request)
	}
	defer tempFile.Close()

	startTime := time.Now()
	e = handler.blobstore.Put(identifier, tempFile)
	handler.metricsService.SendTimingMetric(handler.resourceType+"-cp_to_blobstore-time", time.Since(startTime))

	if e != nil {
		handler.notifyUploadFailed(identifier, e, request)
		if _, noSpaceLeft := e.(*NoSpaceLeftError); noSpaceLeft {
			return handle(e, async, request)
		}
		return handle(errors.Wrapf(e, "Could not upload temporary file to blobstore"), async, request)
	}
	sha1, sha256, e := ShaSums(tempFilename)
	if e != nil {
		handler.notifyUploadFailed(identifier, e, request)
		return handle(errors.Wrapf(e, "Could not build sha sums of temporary file '%v'", tempFilename), async, request)
	}
	e = handler.updater.NotifyUploadSucceeded(identifier, hex.EncodeToString(sha1), hex.EncodeToString(sha256))
	if e != nil {
		return handle(errors.Wrapf(e, "Could not notify Cloud Controller about successful upload"), async, request)
	}
	return nil
}

// TODO(pego): find better name for this function
func handle(e error, async bool, request *http.Request) error {
	if async {
		logger.From(request).Errorw("Failure during upload", "error", e)
		return nil
	}
	return e
}

func (handler *ResourceHandler) notifyUploadFailed(identifier string, e error, request *http.Request) {
	notifyErr := handler.updater.NotifyUploadFailed(identifier, e)
	if notifyErr != nil {
		logger.From(request).Errorw("Failed to notifying CC about failed upload.", "error", notifyErr)
	}
}

func ShaSums(filename string) (sha1Sum []byte, sha256Sum []byte, e error) {
	file, e := os.Open(filename)
	if e != nil {
		return nil, nil, errors.WithStack(e)
	}
	defer file.Close()
	sha1Hash := sha1.New()
	sha256Hash := sha256.New()
	_, e = io.Copy(io.MultiWriter(sha1Hash, sha256Hash), file)
	if e != nil {
		return nil, nil, errors.WithStack(e)
	}
	return sha1Hash.Sum(nil), sha256Hash.Sum(nil), nil
}

func handleNotificationError(e error, responseWriter http.ResponseWriter, request *http.Request) (wasError bool) {
	switch e.(type) {
	case *StateForbiddenError:
		responseWriter.WriteHeader(http.StatusBadRequest)
		fprintDescriptionAndCodeAsJSON(responseWriter, "290008", "Cannot update an existing package.")
		return true
	case *NotFoundError:
		responseWriter.WriteHeader(http.StatusNotFound)
		fprintDescriptionAndCodeAsJSON(responseWriter, "10010", e.Error())
		return true
	case error:
		internalServerError(responseWriter, request, e)
		return true
	}
	return false
}

func (handler *ResourceHandler) CopySourceGuid(responseWriter http.ResponseWriter, request *http.Request, params map[string]string) {
	if !HandleBodySizeLimits(responseWriter, request, handler.maxBodySizeLimit) {
		return
	}
	sourceGuid := sourceGuidFrom(request, responseWriter)
	if sourceGuid == "" {
		return // response is already handled in sourceGuidFrom
	}
	e := handler.blobstore.Copy(sourceGuid, params["identifier"])
	// TODO use Clock instead:
	writeResponseBasedOn("", e, responseWriter, request, http.StatusCreated, nil, &responseBody{Guid: params["identifier"], State: "READY", Type: "bits", CreatedAt: time.Now()}, "")
}

func sourceGuidFrom(request *http.Request, responseWriter http.ResponseWriter) string {
	content, e := ioutil.ReadAll(request.Body)
	if e != nil {
		internalServerError(responseWriter, request, e)
		return ""
	}
	var payload struct {
		SourceGuid string `json:"source_guid"`
	}
	e = json.Unmarshal(content, &payload)
	if e != nil {
		badRequest(responseWriter, request, "Body must be valid JSON when request is not multipart/form-data. %+v", e)
		return ""
	}
	return payload.SourceGuid
}

func (handler *ResourceHandler) HeadOrRedirectAsGet(responseWriter http.ResponseWriter, request *http.Request, params map[string]string) {
	redirectLocation, e := handler.blobstore.HeadOrRedirectAsGet(params["identifier"])
	writeResponseBasedOn(redirectLocation, e, responseWriter, request, http.StatusOK, nil, nil, "")
}

func (handler *ResourceHandler) Get(responseWriter http.ResponseWriter, request *http.Request, params map[string]string) {
	body, redirectLocation, e := handler.blobstore.GetOrRedirect(params["identifier"])
	writeResponseBasedOn(redirectLocation, e, responseWriter, request, http.StatusOK, body, nil, request.Header.Get("If-None-Modify"))
}

func (handler *ResourceHandler) Delete(responseWriter http.ResponseWriter, request *http.Request, params map[string]string) {
	// TODO nothing should be S3 specific here
	// this check is needed, because S3 does not return a NotFound on a Delete request:
	exists, e := handler.blobstore.Exists(params["identifier"])
	if e != nil {
		internalServerError(responseWriter, request, e)
		return
	}
	if !exists {
		responseWriter.WriteHeader(http.StatusNotFound)
		return
	}
	e = handler.blobstore.Delete(params["identifier"])
	writeResponseBasedOn("", e, responseWriter, request, http.StatusNoContent, nil, nil, "")
}

func (handler *ResourceHandler) DeleteDir(responseWriter http.ResponseWriter, request *http.Request, params map[string]string) {
	e := handler.blobstore.DeleteDir(params["identifier"])
	switch e.(type) {
	case *NotFoundError:
		responseWriter.WriteHeader(http.StatusNoContent)
		return
	}
	writeResponseBasedOn("", e, responseWriter, request, http.StatusNoContent, nil, nil, "")
}

var emptyReader = ioutil.NopCloser(bytes.NewReader(nil))

func writeResponseBasedOn(redirectLocation string, e error, responseWriter http.ResponseWriter, request *http.Request, statusCode int, body io.ReadCloser, jsonBody *responseBody, ifNoneModify string) {
	switch e.(type) {
	case *NotFoundError:
		responseWriter.WriteHeader(http.StatusNotFound)
		return
	case *NoSpaceLeftError:
		http.Error(responseWriter, descriptionAndCodeAsJSON("500000", "Request Entity Too Large"), http.StatusInsufficientStorage)
		return
	case error:
		internalServerError(responseWriter, request, e)
		return
	}
	if redirectLocation != "" {
		redirect(responseWriter, redirectLocation)
		return
	}
	if body != nil {
		defer body.Close()
		var buffer bytes.Buffer
		sha := sha1.New()
		_, e := io.Copy(io.MultiWriter(&buffer, sha), body)
		if e != nil {
			internalServerError(responseWriter, request, e)
			return
		}
		eTag := hex.EncodeToString(sha.Sum(nil))
		logger.From(request).Debugw("Cache check", "if-none-modify", ifNoneModify, "etag", eTag)
		responseWriter.Header().Set("ETag", eTag)
		if ifNoneModify == eTag {
			responseWriter.WriteHeader(http.StatusNotModified)
			return
		}
		responseWriter.WriteHeader(statusCode)
		io.Copy(responseWriter, &buffer)
		return
	}
	if jsonBody != nil {
		respBody, marshallingErr := json.Marshal(jsonBody)
		if marshallingErr != nil {
			internalServerError(responseWriter, request, marshallingErr)
			return
		}
		responseWriter.WriteHeader(statusCode)
		responseWriter.Write(respBody)
		return
	}
	responseWriter.WriteHeader(statusCode)
}

func redirect(responseWriter http.ResponseWriter, redirectLocation string) {
	responseWriter.Header().Set("Location", redirectLocation)
	responseWriter.WriteHeader(http.StatusFound)
}

func internalServerError(responseWriter http.ResponseWriter, request *http.Request, e error) {
	logger.From(request).Errorw("Internal Server Error.", "error", fmt.Sprintf("%+v", e))
	responseWriter.WriteHeader(http.StatusInternalServerError)
}

func badRequest(responseWriter http.ResponseWriter, request *http.Request, message string, args ...interface{}) {
	responseBody := fmt.Sprintf(message, args...)
	logger.From(request).Infow("Bad request", "body", responseBody)
	responseWriter.WriteHeader(http.StatusBadRequest)
	fprintDescriptionAndCodeAsJSON(responseWriter, "290003", message, args...)
}

func fprintDescriptionAndCodeAsJSON(responseWriter http.ResponseWriter, code string, description string, a ...interface{}) {
	fmt.Fprintf(responseWriter, descriptionAndCodeAsJSON(code, description, a...))
}

func descriptionAndCodeAsJSON(code string, description string, a ...interface{}) string {
	return fmt.Sprintf(`{"description":"%v","code":%v}`, fmt.Sprintf(description, a...), code)
}
