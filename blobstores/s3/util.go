package s3

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

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

func isS3NotFoundError(e error) bool {
	if ae, isAwsErr := e.(awserr.Error); isAwsErr {
		if ae.Code() == "NoSuchBucket" || ae.Code() == "NoSuchKey" || ae.Code() == "NotFound" {
			return true
		}
	}
	return false
}
