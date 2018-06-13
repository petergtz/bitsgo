package s3_test

import (
	"testing"
	"time"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"github.com/petergtz/bitsgo/internal/blobstores/decorator"
	. "github.com/petergtz/bitsgo/internal/blobstores/s3"
	"github.com/petergtz/bitsgo/internal/config"
)

func TestS3Blobstore(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "S3Blobstore")
}

var _ = Describe("Signing URLs", func() {
	It("Can create pre-signed URLs for S3", func() {
		signer := decorator.ForResourceSignerWithPathPartitioning(NewBlobstore(
			config.S3BlobstoreConfig{"mybucket", "MY-Key_ID", "dummy", "us-east-1", ""}))

		signedURL := signer.Sign("myresource", "get", time.Now())

		Expect(signedURL).To(SatisfyAll(
			ContainSubstring("https://mybucket.s3.amazonaws.com/my/re/myresource"),
			ContainSubstring("X-Amz-Algorithm="),
			ContainSubstring("X-Amz-Credential=MY-Key_ID"),
			ContainSubstring("X-Amz-Date="),
			ContainSubstring("X-Amz-Expires="),
			ContainSubstring("X-Amz-Signature="),
		))
	})
})
