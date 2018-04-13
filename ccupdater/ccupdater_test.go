package ccupdater_test

import (
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/petergtz/bitsgo/ccupdater"
	. "github.com/petergtz/bitsgo/ccupdater/matchers"
	. "github.com/petergtz/pegomock"
)

var _ = Describe("CCUpdater", func() {
	It("works", func() {
		httpClient := NewMockHttpClient()
		updater := NewCCUpdaterWithHttpClient("http://example.com/some/endpoint", "PATCH", httpClient)

		When(httpClient.Do(AnyPtrToHttpRequest())).ThenReturn(&http.Response{}, nil)

		e := updater.NotifyProcessingUpload("abc")

		Expect(e).NotTo(HaveOccurred())

		request := httpClient.VerifyWasCalledOnce().Do(AnyPtrToHttpRequest()).GetCapturedArguments()
		Expect(request.Method).To(Equal("PATCH"))
		Expect(request.URL.String()).To(Equal("http://example.com/some/endpoint/abc"))
	})
})
