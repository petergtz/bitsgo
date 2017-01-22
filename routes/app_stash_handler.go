package routes

import (
	"archive/zip"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/petergtz/bitsgo/logger"
	"github.com/pkg/errors"
	"github.com/uber-go/zap"
)

type AppStashHandler struct {
	blobstore Blobstore
}

func (handler *AppStashHandler) PostMatches(responseWriter http.ResponseWriter, request *http.Request) {
	body, e := ioutil.ReadAll(request.Body)
	if e != nil {
		internalServerError(responseWriter, e)
		return
	}
	var sha1s []struct {
		Sha1 string
		Size int
	}
	e = json.Unmarshal(body, &sha1s)
	if e != nil {
		log.Printf("Invalid body %s\n\n%v", body, e)
		responseWriter.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprintf(responseWriter, "Invalid body %s", body)
		return
	}
	if len(sha1s) == 0 {
		// TODO improve messages
		log.Printf("Empty list %s\n\n%v", body, e)
		responseWriter.WriteHeader(http.StatusUnprocessableEntity)
		fprintDescriptionAsJSON(responseWriter, "The request is semantically invalid: must be a non-empty array.")
		return
	}
	responseSha1 := []map[string]string{}
	for _, entry := range sha1s {
		exists, e := handler.blobstore.Exists(entry.Sha1)
		if e != nil {
			internalServerError(responseWriter, e)
			return
		}
		if exists {
			responseSha1 = append(responseSha1, map[string]string{"sha1": entry.Sha1})
		}
	}
	response, e := json.Marshal(&responseSha1)
	if e != nil {
		internalServerError(responseWriter, e)
		return
	}
	responseWriter.Write(response)
}

func (handler *AppStashHandler) PostEntries(responseWriter http.ResponseWriter, request *http.Request) {
	uploadedFile, _, e := request.FormFile("application")
	if e != nil {
		badRequest(responseWriter, "Could not retrieve 'application' form parameter")
		return
	}
	defer uploadedFile.Close()

	tempZipFile, e := ioutil.TempFile("", "")
	if e != nil {
		internalServerError(responseWriter, e)
		return
	}
	defer os.Remove(tempZipFile.Name())
	defer tempZipFile.Close()

	_, e = io.Copy(tempZipFile, uploadedFile)
	if e != nil {
		internalServerError(responseWriter, e)
		return
	}

	openZipFile, e := zip.OpenReader(tempZipFile.Name())
	if e != nil {
		badRequest(responseWriter, "Bad Request: Not a valid zip file")
		return
	}
	defer openZipFile.Close()

	bundlesPayload := []BundlesPayload{}
	for _, zipFileEntry := range openZipFile.File {
		if !zipFileEntry.FileInfo().Mode().IsRegular() {
			continue
		}
		sha, e := copyTo(handler.blobstore, zipFileEntry)
		if e != nil {
			internalServerError(responseWriter, e)
			return
		}
		logger.Log.Debug("Filemode in zip File Entry",
			zap.String("filemode", zipFileEntry.FileInfo().Mode().String()),
		)
		if e != nil {
			internalServerError(responseWriter, e)
			return
		}
		bundlesPayload = append(bundlesPayload, BundlesPayload{
			Sha1: sha,
			Fn:   zipFileEntry.Name,
			Mode: strconv.FormatInt(int64(zipFileEntry.FileInfo().Mode()), 8),
		})
	}
	receipt, e := json.Marshal(bundlesPayload)
	if e != nil {
		internalServerError(responseWriter, e)
		return
	}
	responseWriter.WriteHeader(http.StatusCreated)
	responseWriter.Write(receipt)
}

func copyTo(blobstore Blobstore, zipFileEntry *zip.File) (sha string, err error) {
	unzippedReader, e := zipFileEntry.Open()
	if e != nil {
		return "", errors.WithStack(e)
	}
	defer unzippedReader.Close()

	tempZipEntryFile, e := ioutil.TempFile("", filepath.Base(zipFileEntry.Name))
	if e != nil {
		return "", errors.WithStack(e)
	}
	defer os.Remove(tempZipEntryFile.Name())
	defer tempZipEntryFile.Close()

	sha, e = copyCalculatingSha(tempZipEntryFile, unzippedReader)
	if e != nil {
		return "", errors.WithStack(e)
	}

	entryFileRead, e := os.Open(tempZipEntryFile.Name())
	if e != nil {
		return "", errors.WithStack(e)
	}
	defer entryFileRead.Close()

	// TODO: this assumes no redirect on PUTs. Is that always true?
	_, e = blobstore.Put(sha, entryFileRead)
	if e != nil {
		return "", errors.WithStack(e)
	}

	return
}

func copyCalculatingSha(writer io.Writer, reader io.Reader) (sha string, e error) {
	checkSum := sha1.New()
	multiWriter := io.MultiWriter(writer, checkSum)

	_, e = io.Copy(multiWriter, reader)
	if e != nil {
		return "", fmt.Errorf("error copying. Caused by: %v", e)
	}

	return fmt.Sprintf("%x", checkSum.Sum(nil)), nil
}

type BundlesPayload struct {
	Sha1 string `json:"sha1"`
	Fn   string `json:"fn"`
	Mode string `json:"mode"`
}

func (handler *AppStashHandler) PostBundles(responseWriter http.ResponseWriter, request *http.Request) {
	body, e := ioutil.ReadAll(request.Body)
	if e != nil {
		internalServerError(responseWriter, e)
		return
	}

	var bundlesPayload []BundlesPayload
	e = json.Unmarshal(body, &bundlesPayload)
	if e != nil {
		log.Printf("Invalid body %s", body)
		responseWriter.WriteHeader(http.StatusUnprocessableEntity)
		fprintDescriptionAsJSON(responseWriter, "Invalid body %s", body)
		return
	}

	if isMissing, key := anyKeyMissingIn(bundlesPayload); isMissing {
		responseWriter.WriteHeader(http.StatusUnprocessableEntity)
		fprintDescriptionAsJSON(responseWriter, "The request is semantically invalid: key `%v` missing or empty", key)
		return
	}

	tempZipFilename, e := handler.createTempZipFileFrom(bundlesPayload)
	if e != nil {
		if notFoundError, ok := e.(*NotFoundError); ok {
			responseWriter.WriteHeader(http.StatusNotFound)
			fprintDescriptionAsJSON(responseWriter, "%v not found", notFoundError.Error())
			return
		}
		internalServerError(responseWriter, e)
		return
	}
	defer os.Remove(tempZipFilename)

	tempZipFile, e := os.Open(tempZipFilename)
	if e != nil {
		internalServerError(responseWriter, e)
		return
	}
	defer tempZipFile.Close()

	_, e = io.Copy(responseWriter, tempZipFile)
	if e != nil {
		internalServerError(responseWriter, e)
		return
	}
}

func fprintDescriptionAsJSON(responseWriter http.ResponseWriter, description string, a ...interface{}) {
	fmt.Fprintf(responseWriter, `{"description":"%v"}`, fmt.Sprintf(description, a...))
}

func anyKeyMissingIn(bundlesPayload []BundlesPayload) (bool, string) {
	for _, entry := range bundlesPayload {
		if entry.Sha1 == "" {
			return true, "sha1"
		}
		if entry.Fn == "" {
			return true, "fn"
		}
	}
	return false, ""
}

func (handler *AppStashHandler) createTempZipFileFrom(bundlesPayload []BundlesPayload) (tempFilename string, err error) {
	tempFile, e := ioutil.TempFile("", "bundles")
	if e != nil {
		return "", e
	}
	defer tempFile.Close()
	zipWriter := zip.NewWriter(tempFile)
	for _, entry := range bundlesPayload {
		zipEntry, e := zipWriter.CreateHeader(zipEntryHeader(entry.Fn, fileModeFrom(entry.Mode)))
		if e != nil {
			return "", e
		}

		// TODO this assumes no redirects. Probably app_stash should have its own interface for blobstore that expresses no redirects
		b, _, e := handler.blobstore.Get(entry.Sha1)
		if e != nil {
			if _, ok := e.(*NotFoundError); ok {
				return "", NewNotFoundErrorWithMessage(entry.Sha1)
			}
			return "", e
		}
		defer b.Close()

		_, e = io.Copy(zipEntry, b)
		if e != nil {
			return "", e
		}
	}
	zipWriter.Close()
	return tempFile.Name(), nil
}

func fileModeFrom(s string) os.FileMode {
	mode, e := strconv.ParseInt(s, 8, 32)
	if e != nil {
		return 0744
	}
	return os.FileMode(mode)
}

func zipEntryHeader(name string, mode os.FileMode) *zip.FileHeader {
	header := &zip.FileHeader{
		Name:   name,
		Method: zip.Deflate,
	}
	header.SetMode(mode)
	return header
}
