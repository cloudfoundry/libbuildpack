package libbuildpack_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	httpmock "github.com/jarcoal/httpmock"

	"testing"
)

var _ = BeforeSuite(func() {
	httpmock.Activate()
})

var _ = AfterSuite(func() {
	httpmock.DeactivateAndReset()
})

func TestLibBuildpackTest(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "LibBuildpackTest Suite")
}
