package s3_blobstore

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gorilla/mux"
)

// const DefaultS3Region = "us-east-1"
const DefaultS3Region = "eu-west-1"

func newS3Client(region string, accessKeyID string, secretAccessKey string) *s3.S3 {
	session, e := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""),
	})
	if e != nil {
		panic(e)
	}
	return s3.New(session)
}

type SignS3UrlHandler struct {
	s3Client *s3.S3
	bucket   string
}

func NewSignS3UrlHandler(bucket string, accessKeyID, secretAccessKey string) *SignS3UrlHandler {
	return &SignS3UrlHandler{
		s3Client: newS3Client(DefaultS3Region, accessKeyID, secretAccessKey),
		bucket:   bucket,
	}
}

func (handler *SignS3UrlHandler) Sign(responseWriter http.ResponseWriter, r *http.Request) {
	var request *request.Request
	if r.URL.Query().Get("verb") == "put" {
		request, _ = handler.s3Client.PutObjectRequest(&s3.PutObjectInput{
			Bucket: aws.String(handler.bucket),
			// TODO this shouldn't use mux.Vars directly. Instead, this should be refactored.
			Key: aws.String(pathFor(mux.Vars(r)["guid"])),
		})
	} else {
		request, _ = handler.s3Client.GetObjectRequest(&s3.GetObjectInput{
			Bucket: aws.String(handler.bucket),
			Key:    aws.String(pathFor(mux.Vars(r)["guid"])),
		})
	}
	// TODO what expiration duration should we use?
	signedURL, e := request.Presign(time.Hour)
	if e != nil {
		panic(e)
	}
	log.Printf("Signed URL (verb=%v): %v", r.URL.Query().Get("verb"), signedURL)
	fmt.Fprint(responseWriter, signedURL)
}

// TODO this is a duplicate from resource_handler.go and should be refactored.
func pathFor(identifier string) string {
	return fmt.Sprintf("/%s/%s/%s", identifier[0:2], identifier[2:4], identifier)
}
