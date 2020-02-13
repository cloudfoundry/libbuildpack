package glow_test

import (
	"errors"
	"fmt"

	"github.com/cloudfoundry/libbuildpack/cutlass/glow"
	"github.com/cloudfoundry/libbuildpack/cutlass/glow/fakes"
	"github.com/cloudfoundry/packit/pexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CLI", func() {
	var (
		cli        glow.CLI
		executable *fakes.Executable
	)

	BeforeEach(func() {
		executable = &fakes.Executable{}
		executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
			fmt.Fprintf(execution.Stdout, "some-stdout")
			fmt.Fprintf(execution.Stderr, "some-stderr")

			return nil
		}

		cli = glow.NewCLI(executable)
	})

	Describe("Package", func() {
		It("calls the package subcommand with the correct arguments", func() {
			stdout, stderr, err := cli.Package("some-dir", "some-stack", glow.PackageOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(stdout).To(Equal("some-stdout"))
			Expect(stderr).To(Equal("some-stderr"))

			Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{"package", "-stack", "some-stack"}))
			Expect(executable.ExecuteCall.Receives.Execution.Dir).To(Equal("some-dir"))
		})

		It("calls the package subcommand with the -version flag", func() {
			stdout, stderr, err := cli.Package("some-dir", "some-stack", glow.PackageOptions{
				Version: "some-version",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(stdout).To(Equal("some-stdout"))
			Expect(stderr).To(Equal("some-stderr"))

			Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{"package", "-stack", "some-stack", "-version", "some-version"}))
			Expect(executable.ExecuteCall.Receives.Execution.Dir).To(Equal("some-dir"))
		})

		It("calls the package subcommand with the -manifestpath flag", func() {
			stdout, stderr, err := cli.Package("some-dir", "some-stack", glow.PackageOptions{
				ManifestPath: "some-path",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(stdout).To(Equal("some-stdout"))
			Expect(stderr).To(Equal("some-stderr"))

			Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{"package", "-stack", "some-stack", "-manifestpath", "some-path"}))
			Expect(executable.ExecuteCall.Receives.Execution.Dir).To(Equal("some-dir"))
		})

		It("calls the package subcommand with the -dev flag", func() {
			stdout, stderr, err := cli.Package("some-dir", "some-stack", glow.PackageOptions{
				Dev: true,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(stdout).To(Equal("some-stdout"))
			Expect(stderr).To(Equal("some-stderr"))

			Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{"package", "-stack", "some-stack", "-dev"}))
			Expect(executable.ExecuteCall.Receives.Execution.Dir).To(Equal("some-dir"))
		})

		It("calls the package subcommand with the -cached flag", func() {
			stdout, stderr, err := cli.Package("some-dir", "some-stack", glow.PackageOptions{
				Cached: true,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(stdout).To(Equal("some-stdout"))
			Expect(stderr).To(Equal("some-stderr"))

			Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{"package", "-stack", "some-stack", "-cached"}))
			Expect(executable.ExecuteCall.Receives.Execution.Dir).To(Equal("some-dir"))
		})

		Context("failure cases", func() {
			Context("when the executable fails", func() {
				BeforeEach(func() {
					executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
						fmt.Fprintf(execution.Stdout, "some-stdout")
						fmt.Fprintf(execution.Stderr, "some-stderr")
						return errors.New("failed to execute")
					}
				})

				It("returns the error and stdout and stderr", func() {
					stdout, stderr, err := cli.Package("some-dir", "some-stack", glow.PackageOptions{})
					Expect(err).To(MatchError("failed to execute"))
					Expect(stdout).To(Equal("some-stdout"))
					Expect(stderr).To(Equal("some-stderr"))
				})
			})
		})
	})
})
