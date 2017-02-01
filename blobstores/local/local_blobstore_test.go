package local_test

import (
	"bytes"
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/petergtz/bitsgo/blobstores/local"
	. "github.com/petergtz/pegomock"
)

var _ = Describe("Blobstore", func() {
	Describe("Put", func() {
		Context("Creating a file returns an error", func() {
			It("returns an error and closes other files properly", func() {
				fs := NewMockFs()
				blobstore := local.NewBlobstoreWithFs(fs)

				When(fs.Create(AnyString())).ThenReturn(nil, errors.New("some error"))

				redirectLocatin, e := blobstore.Put("some path", bytes.NewReader([]byte("content")))

				Expect(redirectLocatin).To(BeEmpty())
				Expect(e).To(HaveOccurred())
				fmt.Println(e)
			})
		})

	})
})
