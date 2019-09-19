package libbuildpack_test

import (
	"testing"

	"github.com/jarcoal/httpmock"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = BeforeSuite(func() {
	httpmock.Activate()
})

var _ = AfterSuite(func() {
	httpmock.DeactivateAndReset()
})

func TestLibBuildpackTest(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "libbuildpack")
}
