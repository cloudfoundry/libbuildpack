package execution_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestExecution(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "cutlass/execution")
}

var (
	fakeCLI      string
	existingPath string
)

var _ = BeforeSuite(func() {
	var err error
	fakeCLI, err = gexec.Build("github.com/cloudfoundry/libbuildpack/cutlass/execution/fakes/some-executable")
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})

var _ = BeforeEach(func() {
	existingPath = os.Getenv("PATH")
	os.Setenv("PATH", filepath.Dir(fakeCLI))
})

var _ = AfterEach(func() {
	os.Setenv("PATH", existingPath)
})
