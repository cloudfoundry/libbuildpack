package docker_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry/libbuildpack/cutlass/docker"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DockerExecutable", func() {
	var (
		executable docker.DockerExecutable
		tmpDir     string
	)

	BeforeEach(func() {
		var err error
		tmpDir, err = ioutil.TempDir("", "docker-executable")
		Expect(err).NotTo(HaveOccurred())

		tmpDir, err = filepath.EvalSymlinks(tmpDir)
		Expect(err).NotTo(HaveOccurred())

		logger := lager.NewLogger("cutlass")

		executable = docker.NewDockerExecutable(logger)
	})

	AfterEach(func() {
		Expect(os.RemoveAll(tmpDir)).To(Succeed())
	})

	Describe("Execute", func() {
		It("executes the given arguments against the executable", func() {
			stdout, stderr, err := executable.Execute(docker.ExecuteOptions{Dir: tmpDir}, "something")
			Expect(err).NotTo(HaveOccurred())
			Expect(stdout).To(ContainSubstring("Output on stdout"))
			Expect(stderr).To(ContainSubstring("Output on stderr"))

			Expect(stdout).To(ContainSubstring("Arguments: [docker something]"))
			Expect(stdout).To(ContainSubstring(fmt.Sprintf("PWD: %s", tmpDir)))
		})

		Context("when docker is not on the path", func() {
			BeforeEach(func() {
				os.Unsetenv("PATH")
			})

			It("executes the given arguments against the executable", func() {
				_, _, err := executable.Execute(docker.ExecuteOptions{}, "something")
				Expect(err).To(MatchError("exec: \"docker\": executable file not found in $PATH"))
			})
		})
	})
})
