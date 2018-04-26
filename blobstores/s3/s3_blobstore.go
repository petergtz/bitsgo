package s3

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/petergtz/bitsgo"
	"github.com/petergtz/bitsgo/blobstores/validate"
	"github.com/petergtz/bitsgo/config"
	"github.com/petergtz/bitsgo/logger"
	"github.com/pkg/errors"
	// v4signer "github.com/aws/aws-sdk-go/aws/signer/v4"
	v2signer "github.com/aws/aws-sdk-go/private/signer/v2"
)

type Blobstore struct {
	s3Client        *s3.S3
	S3Client        *s3.S3
	bucket          string
	accessKeyID     string
	secretAccessKey string
}

func NewBlobstore(config config.S3BlobstoreConfig) *Blobstore {
	validate.NotEmpty(config.AccessKeyID)
	validate.NotEmpty(config.Bucket)
	// validate.NotEmpty(config.Region)
	validate.NotEmpty(config.SecretAccessKey)

	return &Blobstore{
		s3Client:        newS3Client(config.Region, config.AccessKeyID, config.SecretAccessKey, config.Host),
		S3Client:        newS3Client(config.Region, config.AccessKeyID, config.SecretAccessKey, config.Host),
		bucket:          config.Bucket,
		accessKeyID:     config.AccessKeyID,
		secretAccessKey: config.SecretAccessKey,
	}
}

func (blobstore *Blobstore) Exists(path string) (bool, error) {
	_, e := blobstore.s3Client.HeadObject(&s3.HeadObjectInput{
		Bucket: &blobstore.bucket,
		Key:    &path,
	})
	if e != nil {
		if isS3NotFoundError(e) {
			return false, nil
		}
		return false, errors.Wrapf(e, "Failed to check for %v/%v", blobstore.bucket, path)
	}
	return true, nil
}

func (blobstore *Blobstore) HeadOrRedirectAsGet(path string) (redirectLocation string, err error) {
	request, _ := blobstore.s3Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: &blobstore.bucket,
		Key:    &path,
	})
	return signedURLFrom(request, blobstore.bucket, path)
}

func (blobstore *Blobstore) Get(path string) (body io.ReadCloser, err error) {
	logger.Log.Debugw("Get from S3", "bucket", blobstore.bucket, "path", path)
	output, e := blobstore.s3Client.GetObject(&s3.GetObjectInput{
		Bucket: &blobstore.bucket,
		Key:    &path,
	})
	if e != nil {
		if isS3NotFoundError(e) {
			return nil, bitsgo.NewNotFoundError()
		}
		return nil, errors.Wrapf(e, "Path %v", path)
	}
	return output.Body, nil
}

func (blobstore *Blobstore) GetOrRedirect(path string) (body io.ReadCloser, redirectLocation string, err error) {
	request, _ := blobstore.s3Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: &blobstore.bucket,
		Key:    &path,
	})
	signedUrl, e := signedURLFrom(request, blobstore.bucket, path)
	return nil, signedUrl, e
}

func (blobstore *Blobstore) Put(path string, src io.ReadSeeker) error {
	logger.Log.Debugw("Put to S3", "bucket", blobstore.bucket, "path", path)
	_, e := blobstore.s3Client.PutObject(&s3.PutObjectInput{
		Bucket: &blobstore.bucket,
		Key:    &path,
		Body:   src,
	})
	if e != nil {
		return errors.Wrapf(e, "Path %v", path)
	}
	return nil
}

const timeFormat = "2006-01-02T15:04:05"

func (blobstore *Blobstore) SignedURLFromX(req *request.Request, bucket, path string) (string, error) {
	req.Build()

	q := req.HTTPRequest.URL.Query()

	q.Set("AWSAccessKeyId", blobstore.accessKeyID)
	q.Set("Expires", fmt.Sprintf("%v", "1524753676"))
	q.Set("SignatureVersion", "2")
	q.Set("SignatureMethod", "HmacSHA256")
	q.Set("Timestamp", time.Now().Format(timeFormat))
	q.Set("Version", "2009-03-31")

	req.HTTPRequest.URL.RawQuery = q.Encode()

	signature := req.HTTPRequest.Method + "\n" +
		req.HTTPRequest.URL.Host + "\n" +
		req.HTTPRequest.URL.Path + "\n" +
		req.HTTPRequest.URL.Query().Encode()

	fmt.Println(signature)

	hash := hmac.New(sha256.New, []byte(blobstore.secretAccessKey))
	hash.Write([]byte(signature))

	q.Set("Signature", base64.StdEncoding.EncodeToString(hash.Sum(nil)))
	q.Set("GoogleAccessId", blobstore.accessKeyID)
	q.Del("AWSAccessKeyId")
	req.HTTPRequest.URL.RawQuery = q.Encode()

	signed := req.HTTPRequest.URL.String()

	return signed, nil
}

func signedURLFrom(req *request.Request, bucket, path string) (string, error) {
	// signedURL2, _ := req.Presign(time.Hour)
	req.Build()

	fmt.Printf("XXX unsigned %+v\n", req.HTTPRequest.URL.String())
	// fmt.Printf("XXX Original %v\n", signedURL2)
	// req.NotHoist = false

	// req.ExpireTime = time.Hour
	// req.Sign()
	v2signer.SignSDKRequest(req)
	fmt.Printf("XXX V2 %v\n", req.HTTPRequest.URL.String())
	// signedURL, e := req.Presign(time.Hour)
	if req.Error != nil {
		panic(req.Error)
	}

	signedURL := req.HTTPRequest.URL.String()
	// if e != nil {
	// 	return "", errors.Wrapf(e, "Bucket/Path %v/%v", bucket, path)
	// }

	signedURL = strings.Replace(signedURL, "AWSAccessKeyId", "GoogleAccessId", -1)
	fmt.Printf("XXX %v\n", signedURL)
	return signedURL, nil

}

func (blobstore *Blobstore) Copy(src, dest string) error {
	logger.Log.Debugw("Copy in S3", "bucket", blobstore.bucket, "src", src, "dest", dest)
	_, e := blobstore.s3Client.CopyObject(&s3.CopyObjectInput{
		Key:        &dest,
		CopySource: aws.String(blobstore.bucket + "/" + src),
		Bucket:     &blobstore.bucket,
	})
	if e != nil {
		if isS3NotFoundError(e) {
			return bitsgo.NewNotFoundError()
		}
		return errors.Wrapf(e, "Error while trying to copy src %v to dest %v in bucket %v", src, dest, blobstore.bucket)
	}
	return nil
}

func (blobstore *Blobstore) Delete(path string) error {
	_, e := blobstore.s3Client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: &blobstore.bucket,
		Key:    &path,
	})
	if e != nil {
		if isS3NotFoundError(e) {
			return bitsgo.NewNotFoundError()
		}
		return errors.Wrapf(e, "Path %v", path)
	}
	return nil
}

func (blobstore *Blobstore) DeleteDir(prefix string) error {
	deletionErrs := []error{}
	e := blobstore.s3Client.ListObjectsPages(
		&s3.ListObjectsInput{
			Bucket: &blobstore.bucket,
			Prefix: &prefix,
		},
		func(p *s3.ListObjectsOutput, lastPage bool) (shouldContinue bool) {
			for _, object := range p.Contents {
				e := blobstore.Delete(*object.Key)
				if e != nil {
					if _, isNotFoundError := e.(*bitsgo.NotFoundError); !isNotFoundError {
						deletionErrs = append(deletionErrs, e)
					}
				}
			}
			return true
		})
	if e != nil {
		return errors.Wrapf(e, "Prefix %v, errors from deleting: %v", prefix, deletionErrs)
	}
	if len(deletionErrs) != 0 {
		return errors.Errorf("Prefix %v, errors from deleting: %v", prefix, deletionErrs)
	}
	return nil
}

func (signer *Blobstore) Sign(resource string, method string, expirationTime time.Time) (signedURL string) {
	var request *request.Request
	switch strings.ToLower(method) {
	case "put":
		request, _ = signer.s3Client.PutObjectRequest(&s3.PutObjectInput{
			Bucket: aws.String(signer.bucket),
			Key:    aws.String(resource),
		})
	case "get":
		request, _ = signer.s3Client.GetObjectRequest(&s3.GetObjectInput{
			Bucket: aws.String(signer.bucket),
			Key:    aws.String(resource),
		})
	default:
		panic("The only supported methods are 'put' and 'get'. But got '" + method + "'")
	}
	// TODO use clock
	signedURL, e := request.Presign(expirationTime.Sub(time.Now()))
	if e != nil {
		panic(e)
	}
	logger.Log.Debugw("Signed URL", "verb", method, "signed-url", signedURL)
	return
}
