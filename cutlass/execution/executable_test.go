package execution_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry/libbuildpack/cutlass/execution"
	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Executable", func() {
	var (
		executable execution.Executable
		tmpDir     string
	)

	BeforeEach(func() {
		var err error
		tmpDir, err = ioutil.TempDir("", "cnb2cf-executable")
		Expect(err).NotTo(HaveOccurred())

		tmpDir, err = filepath.EvalSymlinks(tmpDir)
		Expect(err).NotTo(HaveOccurred())

		logger := lager.NewLogger("cutlass")

		executable = execution.NewExecutable("some-executable", logger)
	})

	AfterEach(func() {
		Expect(os.RemoveAll(tmpDir)).To(Succeed())
	})

	Describe("Execute", func() {
		It("executes the given arguments against the executable", func() {
			stdout, stderr, err := executable.Execute(execution.Options{Dir: tmpDir}, "something")
			Expect(err).NotTo(HaveOccurred())
			Expect(stdout).To(ContainSubstring("Output on stdout"))
			Expect(stderr).To(ContainSubstring("Output on stderr"))

			Expect(stdout).To(ContainSubstring("Arguments: [some-executable something]"))
			Expect(stdout).To(ContainSubstring(fmt.Sprintf("PWD: %s", tmpDir)))
		})

		Context("when cnb2cf errors", func() {
			var (
				errorCLI string
				path     string
			)

			BeforeEach(func() {
				os.Setenv("PATH", existingPath)

				var err error
				errorCLI, err = gexec.Build("github.com/cloudfoundry/libbuildpack/cutlass/execution/fakes/some-executable", "-ldflags", "-X main.fail=true")
				Expect(err).NotTo(HaveOccurred())

				path = os.Getenv("PATH")
				os.Setenv("PATH", filepath.Dir(errorCLI))
			})

			AfterEach(func() {
				os.Setenv("PATH", path)
			})

			It("executes the given arguments against the executable", func() {
				stdout, stderr, err := executable.Execute(execution.Options{}, "something")
				Expect(err).To(MatchError("exit status 1"))
				Expect(stdout).To(ContainSubstring("Error on stdout"))
				Expect(stderr).To(ContainSubstring("Error on stderr"))
			})
		})
	})
})
