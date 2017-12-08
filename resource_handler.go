package bitsgo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/petergtz/bitsgo/logger"
)

type MetricsService interface {
	SendTimingMetric(name string, duration time.Duration)
}

const (
	StateReady            = "ready"
	StateProcessingUpload = "processing_upload"
	StateFailed           = "state_failed"
)

type ResourceStatus struct {
	State  string
	Sha1   string
	Sha256 string
	err    error
}

type Notifier interface {
	Notify(identifier string, status ResourceStatus)
}

type noopNotifier struct{}

func (notifier *noopNotifier) Notify(identifier string, status ResourceStatus) {}

type ResourceHandler struct {
	blobstore        Blobstore
	resourceType     string
	metricsService   MetricsService
	maxBodySizeLimit uint64
	notifier         Notifier
}

func NewResourceHandler(blobstore Blobstore, resourceType string, metricsService MetricsService, maxBodySizeLimit uint64) *ResourceHandler {
	return &ResourceHandler{
		blobstore:        blobstore,
		resourceType:     resourceType,
		metricsService:   metricsService,
		maxBodySizeLimit: maxBodySizeLimit,
		notifier:         new(noopNotifier),
	}
}

func NewPackageHandler(blobstore Blobstore, notifier Notifier, metricsService MetricsService, maxBodySizeLimit uint64) *ResourceHandler {
	return &ResourceHandler{
		blobstore:        blobstore,
		resourceType:     "package",
		metricsService:   metricsService,
		maxBodySizeLimit: maxBodySizeLimit,
		notifier:         notifier,
	}
}

func (handler *ResourceHandler) Put(responseWriter http.ResponseWriter, request *http.Request, params map[string]string) {
	if !HandleBodySizeLimits(responseWriter, request, handler.maxBodySizeLimit) {
		return
	}

	handler.notifier.Notify(params["identifier"], ResourceStatus{State: StateProcessingUpload})

	if strings.Contains(request.Header.Get("Content-Type"), "multipart/form-data") {
		logger.From(request).Debugw("Multipart upload")
		handler.uploadMultipart(responseWriter, request, params["identifier"])
	} else {
		logger.From(request).Debugw("Copy source guid")
		handler.copySourceGuid(responseWriter, request, params["identifier"])
	}
}

func (handler *ResourceHandler) uploadMultipart(responseWriter http.ResponseWriter, request *http.Request, identifier string) {
	file, _, e := request.FormFile(handler.resourceType)
	if e != nil {
		handler.notifier.Notify(identifier, ResourceStatus{State: StateFailed})
		badRequest(responseWriter, "Could not retrieve '%s' form parameter", handler.resourceType)
		return
	}
	defer file.Close()

	startTime := time.Now()
	e = handler.blobstore.Put(identifier, file)
	handler.metricsService.SendTimingMetric(handler.resourceType+"-cp_to_blobstore-time", time.Since(startTime))

	if e != nil {
		handler.notifier.Notify(identifier, ResourceStatus{State: StateFailed, err: e})
	} else {
		handler.notifier.Notify(identifier, ResourceStatus{State: StateReady, Sha1: "", Sha256: ""})
	}

	writeResponseBasedOn("", e, responseWriter, http.StatusCreated, emptyReader)
}

func (handler *ResourceHandler) copySourceGuid(responseWriter http.ResponseWriter, request *http.Request, identifier string) {
	sourceGuid := sourceGuidFrom(request.Body, responseWriter)
	if sourceGuid == "" {
		return // response is already handled in sourceGuidFrom
	}
	e := handler.blobstore.Copy(sourceGuid, identifier)
	writeResponseBasedOn("", e, responseWriter, http.StatusCreated, emptyReader)
}

func sourceGuidFrom(body io.ReadCloser, responseWriter http.ResponseWriter) string {
	content, e := ioutil.ReadAll(body)
	if e != nil {
		internalServerError(responseWriter, e)
		return ""
	}
	var payload struct {
		SourceGuid string `json:"source_guid"`
	}
	e = json.Unmarshal(content, &payload)
	if e != nil {
		badRequest(responseWriter, "Body must be valid JSON when request is not multipart/form-data. %+v", e)
		return ""
	}
	return payload.SourceGuid
}

func (handler *ResourceHandler) Head(responseWriter http.ResponseWriter, request *http.Request, params map[string]string) {
	redirectLocation, e := handler.blobstore.HeadOrRedirectAsGet(params["identifier"])
	writeResponseBasedOn(redirectLocation, e, responseWriter, http.StatusOK, emptyReader)
}

func (handler *ResourceHandler) Get(responseWriter http.ResponseWriter, request *http.Request, params map[string]string) {
	body, redirectLocation, e := handler.blobstore.GetOrRedirect(params["identifier"])
	writeResponseBasedOn(redirectLocation, e, responseWriter, http.StatusOK, body)
}

func (handler *ResourceHandler) Delete(responseWriter http.ResponseWriter, request *http.Request, params map[string]string) {
	// TODO nothing should be S3 specific here
	// this check is needed, because S3 does not return a NotFound on a Delete request:
	exists, e := handler.blobstore.Exists(params["identifier"])
	if e != nil {
		internalServerError(responseWriter, e)
		return
	}
	if !exists {
		responseWriter.WriteHeader(http.StatusNotFound)
		return
	}
	e = handler.blobstore.Delete(params["identifier"])
	writeResponseBasedOn("", e, responseWriter, http.StatusNoContent, emptyReader)
}

func (handler *ResourceHandler) DeleteDir(responseWriter http.ResponseWriter, request *http.Request, params map[string]string) {
	e := handler.blobstore.DeleteDir(params["identifier"])
	switch e.(type) {
	case *NotFoundError:
		responseWriter.WriteHeader(http.StatusNoContent)
		return
	}
	writeResponseBasedOn("", e, responseWriter, http.StatusNoContent, emptyReader)
}

var emptyReader = ioutil.NopCloser(bytes.NewReader(nil))

func writeResponseBasedOn(redirectLocation string, e error, responseWriter http.ResponseWriter, statusCode int, responseReader io.ReadCloser) {
	switch e.(type) {
	case *NotFoundError:
		responseWriter.WriteHeader(http.StatusNotFound)
		return
	case *NoSpaceLeftError:
		responseWriter.WriteHeader(http.StatusInsufficientStorage)
		return
	case error:
		internalServerError(responseWriter, e)
		return
	}
	if redirectLocation != "" {
		redirect(responseWriter, redirectLocation)
		return
	}
	defer responseReader.Close()
	responseWriter.WriteHeader(statusCode)
	io.Copy(responseWriter, responseReader)
}

func redirect(responseWriter http.ResponseWriter, redirectLocation string) {
	responseWriter.Header().Set("Location", redirectLocation)
	responseWriter.WriteHeader(http.StatusFound)
}

func internalServerError(responseWriter http.ResponseWriter, e error) {
	logger.Log.Errorw("Internal Server Error.", "error", fmt.Sprintf("%+v", e))
	responseWriter.WriteHeader(http.StatusInternalServerError)
}

func badRequest(responseWriter http.ResponseWriter, message string, args ...interface{}) {
	responseBody := fmt.Sprintf(message, args...)
	logger.Log.Debugw("Bad rquest", "body", responseBody)
	responseWriter.WriteHeader(http.StatusBadRequest)
	fmt.Fprintf(responseWriter, responseBody)
}
