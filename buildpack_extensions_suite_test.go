package buildpack_test

import (
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	be "github.com/sesmith177/buildpack-extensions"
	"gopkg.in/jarcoal/httpmock.v1"

	"testing"
)

var _ = BeforeSuite(func() {
	httpmock.Activate()
})

var _ = BeforeEach(func() {
	be.Log.SetOutput(ioutil.Discard)
})

var _ = AfterSuite(func() {
	httpmock.DeactivateAndReset()
})

func TestBuildpackExtensionsTest(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "BuildpackExtensionsTest Suite")
}
