package ruby_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	yaml "gopkg.in/yaml.v2"

	"os"

	"os/exec"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	. "github.com/petergtz/bitsgo/blobstores/ruby"
	"github.com/petergtz/bitsgo/routes"
)

type TestConfig struct {
	FogConnection string `yaml:"fog_connection"`
	DirectoryKey  string `yaml:"directory_key"`
	ScriptDir     string `yaml:"script_dir"`
}

func TestRubyBlobstore(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Ruby Blobstore Contract Integration")
}

var _ = Describe("Ruby Blobstore", func() {
	var (
		testConfig TestConfig
		filepath   string
		blobstore  *Blobstore
	)

	BeforeEach(func() {
		filename := os.Getenv("CONFIG")
		if filename == "" {
			fmt.Println("No $CONFIG set. Defaulting to integration_test_config.yml")
			filename = "integration_test_config.yml"
		}
		file, e := os.Open(filename)
		Expect(e).NotTo(HaveOccurred())
		defer file.Close()
		content, e := ioutil.ReadAll(file)
		Expect(e).NotTo(HaveOccurred())
		e = yaml.Unmarshal(content, &testConfig)
		Expect(e).NotTo(HaveOccurred())
		Expect(testConfig.FogConnection).To(HavePrefix("{"))
		fmt.Println(testConfig.FogConnection)
		Expect(testConfig.DirectoryKey).NotTo(BeEmpty())
		Expect(testConfig.FogConnection).NotTo(BeEmpty())
		Expect(testConfig.ScriptDir).NotTo(BeEmpty())

		os.Chdir(testConfig.ScriptDir)
		output, e := exec.Command("bundle").CombinedOutput()
		Expect(e).NotTo(HaveOccurred())

		fmt.Printf("%s", output)

		filepath = fmt.Sprintf("testfile-%v", time.Now())
	})

	Describe("S3NoRedirectBlobStore", func() {
		BeforeEach(func() {
			blobstore = NewBlobstore(testConfig.FogConnection, testConfig.ScriptDir, testConfig.DirectoryKey)
		})

		It("can put and get a resource there", func() {
			redirectLocation, e := blobstore.HeadOrRedirectAsGet(filepath)
			Expect(e).NotTo(HaveOccurred())
			Expect(redirectLocation).NotTo(BeEmpty())
			Expect(http.Get(redirectLocation)).To(HaveStatusCode(http.StatusNotFound))

			body, e := blobstore.Get(filepath)
			Expect(e).To(BeAssignableToTypeOf(&routes.NotFoundError{}))
			Expect(body).To(BeNil())

			e = blobstore.Put(filepath, strings.NewReader("the file content"))

			redirectLocation, e = blobstore.HeadOrRedirectAsGet(filepath)
			Expect(redirectLocation, e).NotTo(BeEmpty())

			body, e = blobstore.Get(filepath)
			Expect(e).NotTo(HaveOccurred())
			Expect(ioutil.ReadAll(body)).To(ContainSubstring("the file content"))

			e = blobstore.Delete(filepath)
			Expect(e).NotTo(HaveOccurred())

			redirectLocation, e = blobstore.HeadOrRedirectAsGet(filepath)
			Expect(e).NotTo(HaveOccurred())
			Expect(redirectLocation).NotTo(BeEmpty())
			Expect(http.Get(redirectLocation)).To(HaveStatusCode(http.StatusNotFound))

			body, e = blobstore.Get(filepath)
			Expect(e).To(BeAssignableToTypeOf(&routes.NotFoundError{}))
			Expect(body).To(BeNil())
		})

		FIt("Can delete a prefix", func() {
			Expect(blobstore.Exists("one")).To(BeFalse())
			Expect(blobstore.Exists("two")).To(BeFalse())

			redirectLocation, e := blobstore.PutOrRedirect("one", strings.NewReader("the file content"))
			fmt.Println(e)
			Expect(redirectLocation, e).To(BeEmpty())
			redirectLocation, e = blobstore.PutOrRedirect("two", strings.NewReader("the file content"))
			Expect(redirectLocation, e).To(BeEmpty())

			Expect(blobstore.Exists("one")).To(BeTrue())
			Expect(blobstore.Exists("two")).To(BeTrue())

			e = blobstore.DeleteDir("")
			Expect(e).NotTo(HaveOccurred())

			Expect(blobstore.Exists("one")).To(BeFalse())
			Expect(blobstore.Exists("two")).To(BeFalse())
		})

		It("Can delete a prefix like in a file tree", func() {
			Expect(blobstore.Exists("dir/one")).To(BeFalse())
			Expect(blobstore.Exists("dir/two")).To(BeFalse())

			redirectLocation, e := blobstore.PutOrRedirect("dir/one", strings.NewReader("the file content"))
			Expect(redirectLocation, e).To(BeEmpty())
			redirectLocation, e = blobstore.PutOrRedirect("dir/two", strings.NewReader("the file content"))
			Expect(redirectLocation, e).To(BeEmpty())

			Expect(blobstore.Exists("dir/one")).To(BeTrue())
			Expect(blobstore.Exists("dir/two")).To(BeTrue())

			e = blobstore.DeleteDir("dir")
			Expect(e).NotTo(HaveOccurred())

			Expect(blobstore.Exists("dir/one")).To(BeFalse())
			Expect(blobstore.Exists("dir/two")).To(BeFalse())
		})

	})

	Describe("S3PureRedirectBlobstore", func() {
		It("can put and get a resource there", func() {
			blobstore := NewBlobstore(testConfig.FogConnection, testConfig.ScriptDir, testConfig.DirectoryKey)

			redirectLocation, e := blobstore.HeadOrRedirectAsGet(filepath)
			Expect(redirectLocation, e).NotTo(BeEmpty())
			// NOTE: our current contract with bits-service-client requires to do a GET request on a URL received from Head()
			Expect(http.Get(redirectLocation)).To(HaveStatusCode(http.StatusNotFound))

			body, redirectLocation, e := blobstore.GetOrRedirect(filepath)
			Expect(redirectLocation, e).NotTo(BeEmpty())
			Expect(body).To(BeNil())
			Expect(http.Get(redirectLocation)).To(HaveStatusCode(http.StatusNotFound))

			redirectLocation, e = blobstore.PutOrRedirect(filepath, strings.NewReader("the file content"))
			Expect(redirectLocation, e).To(BeEmpty())

			redirectLocation, e = blobstore.HeadOrRedirectAsGet(filepath)
			Expect(redirectLocation, e).NotTo(BeEmpty())
			// NOTE: our current contract with bits-service-client requires to do a GET request on a URL received from Head()
			Expect(http.Get(redirectLocation)).To(HaveStatusCode(http.StatusOK))

			body, redirectLocation, e = blobstore.GetOrRedirect(filepath)
			Expect(redirectLocation, e).NotTo(BeEmpty())
			Expect(body).To(BeNil())
			Expect(http.Get(redirectLocation)).To(HaveBodyWithSubstring("the file content"))

			e = blobstore.Delete(filepath)
			Expect(e).NotTo(HaveOccurred())

			redirectLocation, e = blobstore.HeadOrRedirectAsGet(filepath)
			Expect(redirectLocation, e).NotTo(BeEmpty())
			// NOTE: our current contract with bits-service-client requires to do a GET request on a URL received from Head()
			Expect(http.Get(redirectLocation)).To(HaveStatusCode(http.StatusNotFound))

			body, redirectLocation, e = blobstore.GetOrRedirect(filepath)
			Expect(redirectLocation, e).NotTo(BeEmpty())
			Expect(body).To(BeNil())
			Expect(http.Get(redirectLocation)).To(HaveStatusCode(http.StatusNotFound))
		})
	})

})

func HaveBodyWithSubstring(substring string) types.GomegaMatcher {
	return WithTransform(func(response *http.Response) string {
		actualBytes, e := ioutil.ReadAll(response.Body)
		if e != nil {
			panic(e)
		}
		response.Body.Close()
		return string(actualBytes)
	}, Equal(substring))
}

func HaveStatusCode(statusCode int) types.GomegaMatcher {
	return WithTransform(func(response *http.Response) int {
		return response.StatusCode
	}, Equal(statusCode))
}
