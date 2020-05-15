package docker_test

import (
	"errors"
	"fmt"

	"github.com/cloudfoundry/libbuildpack/cutlass/docker"
	"github.com/cloudfoundry/libbuildpack/cutlass/docker/fakes"
	"github.com/paketo-buildpacks/packit/pexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CLI", func() {
	var (
		executable *fakes.Executable
		cli        docker.CLI
	)

	BeforeEach(func() {
		executable = &fakes.Executable{}
		executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
			fmt.Fprintf(execution.Stdout, "some-stdout-output")
			fmt.Fprintf(execution.Stderr, "some-stderr-output")

			return nil
		}

		cli = docker.NewCLI(executable)
	})

	Describe("Build", func() {
		It("executes the build command against the docker cli", func() {
			stdout, stderr, err := cli.Build(docker.BuildOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(stdout).To(Equal("some-stdout-output"))
			Expect(stderr).To(Equal("some-stderr-output"))

			Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{"build", "."}))
		})

		Context("when given the option to remove the build container", func() {
			It("executes the build command with the --rm flag", func() {
				stdout, stderr, err := cli.Build(docker.BuildOptions{
					Remove: true,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(stdout).To(Equal("some-stdout-output"))
				Expect(stderr).To(Equal("some-stderr-output"))

				Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{"build", "--rm", "."}))
			})
		})

		Context("when given the option to not cache the build image", func() {
			It("executes the build command with the --no-cache flag", func() {
				stdout, stderr, err := cli.Build(docker.BuildOptions{
					NoCache: true,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(stdout).To(Equal("some-stdout-output"))
				Expect(stderr).To(Equal("some-stderr-output"))

				Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{"build", "--no-cache", "."}))
			})
		})

		Context("when given the option to tag the build image", func() {
			It("executes the build command with the --tag flag", func() {
				stdout, stderr, err := cli.Build(docker.BuildOptions{
					Tag: "some-tag",
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(stdout).To(Equal("some-stdout-output"))
				Expect(stderr).To(Equal("some-stderr-output"))

				Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{"build", "--tag", "some-tag", "."}))
			})
		})

		Context("when given the option to specify the Dockerfile for the build image", func() {
			It("executes the build command with the --file flag", func() {
				stdout, stderr, err := cli.Build(docker.BuildOptions{
					File: "some-file",
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(stdout).To(Equal("some-stdout-output"))
				Expect(stderr).To(Equal("some-stderr-output"))

				Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{"build", "--file", "some-file", "."}))
			})
		})

		Context("when given the option to specify the context for the build image", func() {
			It("executes the build command with the given context", func() {
				stdout, stderr, err := cli.Build(docker.BuildOptions{
					Context: "some-context",
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(stdout).To(Equal("some-stdout-output"))
				Expect(stderr).To(Equal("some-stderr-output"))

				Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{"build", "some-context"}))
				Expect(executable.ExecuteCall.Receives.Execution.Dir).To(Equal("some-context"))
			})
		})

		Context("failure cases", func() {
			Context("when the executable cannot execute", func() {
				BeforeEach(func() {
					executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
						fmt.Fprintf(execution.Stdout, "some-stdout-output")
						fmt.Fprintf(execution.Stderr, "some-stderr-output")
						return errors.New("failed to execute")
					}
				})

				It("returns an error, but also includes stdout and stderr", func() {
					stdout, stderr, err := cli.Build(docker.BuildOptions{})
					Expect(err).To(MatchError("failed to execute"))
					Expect(stdout).To(Equal("some-stdout-output"))
					Expect(stderr).To(Equal("some-stderr-output"))
				})
			})
		})
	})

	Describe("Run", func() {
		It("executes the run command against the docker cli", func() {
			stdout, stderr, err := cli.Run("some-image", docker.RunOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(stdout).To(Equal("some-stdout-output"))
			Expect(stderr).To(Equal("some-stderr-output"))

			Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{"run", "some-image"}))
		})

		Context("when given the network option to specify the network the container should use", func() {
			It("executes the run command with the --network flag", func() {
				stdout, stderr, err := cli.Run("some-image", docker.RunOptions{
					Network: "some-network",
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(stdout).To(Equal("some-stdout-output"))
				Expect(stderr).To(Equal("some-stderr-output"))
				Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{"run", "--network", "some-network", "some-image"}))
			})
		})

		Context("when given the remove option to have the container removed after it exits", func() {
			It("executes the run command with the --rm flag", func() {
				stdout, stderr, err := cli.Run("some-image", docker.RunOptions{
					Remove: true,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(stdout).To(Equal("some-stdout-output"))
				Expect(stderr).To(Equal("some-stderr-output"))
				Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{"run", "--rm", "some-image"}))
			})
		})

		Context("when given the tty option to have the container attach a tty", func() {
			It("executes the run command with the --tty flag", func() {
				stdout, stderr, err := cli.Run("some-image", docker.RunOptions{
					TTY: true,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(stdout).To(Equal("some-stdout-output"))
				Expect(stderr).To(Equal("some-stderr-output"))
				Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{"run", "--tty", "some-image"}))
			})
		})

		Context("when given the command option to have the container execute a command", func() {
			It("executes the run command with the given command", func() {
				stdout, stderr, err := cli.Run("some-image", docker.RunOptions{
					Command: "some-command",
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(stdout).To(Equal("some-stdout-output"))
				Expect(stderr).To(Equal("some-stderr-output"))
				Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{"run", "some-image", "bash", "-c", "some-command"}))
			})
		})

		Context("failure cases", func() {
			Context("when the executable fails to execute", func() {
				BeforeEach(func() {
					executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
						fmt.Fprintf(execution.Stdout, "some-stdout-output")
						fmt.Fprintf(execution.Stderr, "some-stderr-output")
						return errors.New("failed to execute")
					}
				})

				It("returns an error, but also includes stdout and stderr", func() {
					stdout, stderr, err := cli.Run("some-image", docker.RunOptions{})
					Expect(err).To(MatchError("failed to execute"))
					Expect(stdout).To(Equal("some-stdout-output"))
					Expect(stderr).To(Equal("some-stderr-output"))
				})
			})
		})
	})

	Describe("RemoveImage", func() {
		It("executes the image rm command against the docker cli", func() {
			stdout, stderr, err := cli.RemoveImage("some-image", docker.RemoveImageOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(stdout).To(Equal("some-stdout-output"))
			Expect(stderr).To(Equal("some-stderr-output"))

			Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{"image", "rm", "some-image"}))
		})

		Context("when given the force option", func() {
			It("executes the image rm command with the --force flag", func() {
				stdout, stderr, err := cli.RemoveImage("some-image", docker.RemoveImageOptions{
					Force: true,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(stdout).To(Equal("some-stdout-output"))
				Expect(stderr).To(Equal("some-stderr-output"))

				Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{"image", "rm", "--force", "some-image"}))
			})
		})

		Context("failure cases", func() {
			Context("when the executable fails to execute", func() {
				BeforeEach(func() {
					executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
						fmt.Fprintf(execution.Stdout, "some-stdout-output")
						fmt.Fprintf(execution.Stderr, "some-stderr-output")
						return errors.New("failed to execute")
					}
				})

				It("returns an error, but also includes stdout and stderr", func() {
					stdout, stderr, err := cli.RemoveImage("some-image", docker.RemoveImageOptions{})
					Expect(err).To(MatchError("failed to execute"))
					Expect(stdout).To(Equal("some-stdout-output"))
					Expect(stderr).To(Equal("some-stderr-output"))
				})
			})
		})
	})
})
