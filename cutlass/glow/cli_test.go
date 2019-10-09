package glow_test

import (
	"errors"

	"github.com/cloudfoundry/libbuildpack/cutlass/execution"
	"github.com/cloudfoundry/libbuildpack/cutlass/glow"
	"github.com/cloudfoundry/libbuildpack/cutlass/glow/fakes"

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
		executable.ExecuteCall.Returns.Stdout = "some-stdout"
		executable.ExecuteCall.Returns.Stderr = "some-stderr"

		cli = glow.NewCLI(executable)
	})

	Describe("Package", func() {
		It("calls the package subcommand with the correct arguments", func() {
			stdout, stderr, err := cli.Package("some-dir", "some-stack", glow.PackageOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(stdout).To(Equal("some-stdout"))
			Expect(stderr).To(Equal("some-stderr"))

			Expect(executable.ExecuteCall.Receives.Args).To(Equal([]string{
				"package",
				"-stack", "some-stack",
			}))
			Expect(executable.ExecuteCall.Receives.Options).To(Equal(execution.Options{
				Dir: "some-dir",
			}))
		})

		It("calls the package subcommand with the -version flag", func() {
			stdout, stderr, err := cli.Package("some-dir", "some-stack", glow.PackageOptions{
				Version: "some-version",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(stdout).To(Equal("some-stdout"))
			Expect(stderr).To(Equal("some-stderr"))

			Expect(executable.ExecuteCall.Receives.Args).To(Equal([]string{
				"package",
				"-stack", "some-stack",
				"-version", "some-version",
			}))
			Expect(executable.ExecuteCall.Receives.Options).To(Equal(execution.Options{
				Dir: "some-dir",
			}))
		})

		It("calls the package subcommand with the -manifestpath flag", func() {
			stdout, stderr, err := cli.Package("some-dir", "some-stack", glow.PackageOptions{
				ManifestPath: "some-path",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(stdout).To(Equal("some-stdout"))
			Expect(stderr).To(Equal("some-stderr"))

			Expect(executable.ExecuteCall.Receives.Args).To(Equal([]string{
				"package",
				"-stack", "some-stack",
				"-manifestpath", "some-path",
			}))
			Expect(executable.ExecuteCall.Receives.Options).To(Equal(execution.Options{
				Dir: "some-dir",
			}))
		})

		It("calls the package subcommand with the -dev flag", func() {
			stdout, stderr, err := cli.Package("some-dir", "some-stack", glow.PackageOptions{
				Dev: true,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(stdout).To(Equal("some-stdout"))
			Expect(stderr).To(Equal("some-stderr"))

			Expect(executable.ExecuteCall.Receives.Args).To(Equal([]string{
				"package",
				"-stack", "some-stack",
				"-dev",
			}))
			Expect(executable.ExecuteCall.Receives.Options).To(Equal(execution.Options{
				Dir: "some-dir",
			}))
		})

		It("calls the package subcommand with the -cached flag", func() {
			stdout, stderr, err := cli.Package("some-dir", "some-stack", glow.PackageOptions{
				Cached: true,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(stdout).To(Equal("some-stdout"))
			Expect(stderr).To(Equal("some-stderr"))

			Expect(executable.ExecuteCall.Receives.Args).To(Equal([]string{
				"package",
				"-stack", "some-stack",
				"-cached",
			}))
			Expect(executable.ExecuteCall.Receives.Options).To(Equal(execution.Options{
				Dir: "some-dir",
			}))
		})

		Context("failure cases", func() {
			Context("when the executable fails", func() {
				BeforeEach(func() {
					executable.ExecuteCall.Returns.Err = errors.New("failed to execute")
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
