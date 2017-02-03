package libbuildpack_test

import (
	"io/ioutil"

	bp "github.com/cloudfoundry/libbuildpack"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/jarcoal/httpmock.v1"

	"testing"
)

var _ = BeforeSuite(func() {
	httpmock.Activate()
})

var _ = BeforeEach(func() {
	bp.Log.SetOutput(ioutil.Discard)
})

var _ = AfterSuite(func() {
	httpmock.DeactivateAndReset()
})

func TestLibBuildpackTest(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "LibBuildpackTest Suite")
}
