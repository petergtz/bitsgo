package bitsgo_test

import (
	"net/http/httptest"
	"reflect"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/petergtz/bitsgo"
	"github.com/petergtz/bitsgo/httputil"
	. "github.com/petergtz/pegomock"
)

func AnyTime() (result time.Time) {
	RegisterMatcher(NewAnyMatcher(reflect.TypeOf((*time.Time)(nil)).Elem()))
	return
}

var getSigner = NewMockResourceSigner()
var putSigner = NewMockResourceSigner()

var _ = Describe("SignResourceHandler", func() {
	It("Signs a GET URL", func() {
		When(getSigner.Sign(AnyString(), AnyString(), AnyTime())).ThenReturn("Some get signature")
		recorder := httptest.NewRecorder()
		handler := bitsgo.NewSignResourceHandler(getSigner, putSigner)
		request := httputil.NewRequest("GET", "/foo", nil).Build()

		handler.Sign(recorder, request, map[string]string{"verb": "get", "resource": "bar"})
		Expect(recorder.Code).To(Equal(200))
		Expect(recorder.Body.String()).To(Equal("Some get signature"))
	})
	Context("missing method", func() {
		It("Responds with error code", func() {
			recorder := httptest.NewRecorder()
			handler := bitsgo.NewSignResourceHandler(getSigner, putSigner)
			request := httputil.NewRequest("", "/does", nil).Build()

			handler.Sign(recorder, request, map[string]string{"verb": "put", "resource": "foobar"})
			Expect(recorder.Code).To(Not(Equal(200)))
			Expect(recorder.Body.String()).To(Equal("Missing verb"))
		})
	})

	It("Signs a PUT URL", func() {
		When(putSigner.Sign(AnyString(), AnyString(), AnyTime())).ThenReturn("Some put signature")

		recorder := httptest.NewRecorder()
		handler := bitsgo.NewSignResourceHandler(getSigner, putSigner)
		request := httputil.NewRequest("PUT", "/bar", nil).Build()

		handler.Sign(recorder, request, map[string]string{"verb": "put", "resource": "foobar"})
		Expect(recorder.Code).To(Equal(200))
		Expect(recorder.Body.String()).To(Equal("Some put signature"))
	})
})
