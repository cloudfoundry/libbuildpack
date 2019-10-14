package docker_test

import (
	"errors"

	"github.com/cloudfoundry/libbuildpack/cutlass/docker"
	"github.com/cloudfoundry/libbuildpack/cutlass/docker/fakes"
	"github.com/cloudfoundry/packit"

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
		executable.ExecuteCall.Returns.Stdout = "some-stdout-output"
		executable.ExecuteCall.Returns.Stderr = "some-stderr-output"

		cli = docker.NewCLI(executable)
	})

	Describe("Build", func() {
		It("executes the build command against the docker cli", func() {
			stdout, stderr, err := cli.Build(docker.BuildOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(stdout).To(Equal("some-stdout-output"))
			Expect(stderr).To(Equal("some-stderr-output"))

			Expect(executable.ExecuteCall.Receives.Execution).To(Equal(packit.Execution{
				Args: []string{"build", "."},
			}))
		})

		Context("when given the option to remove the build container", func() {
			It("executes the build command with the --rm flag", func() {
				stdout, stderr, err := cli.Build(docker.BuildOptions{
					Remove: true,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(stdout).To(Equal("some-stdout-output"))
				Expect(stderr).To(Equal("some-stderr-output"))

				Expect(executable.ExecuteCall.Receives.Execution).To(Equal(packit.Execution{
					Args: []string{"build", "--rm", "."},
				}))
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

				Expect(executable.ExecuteCall.Receives.Execution).To(Equal(packit.Execution{
					Args: []string{"build", "--no-cache", "."},
				}))
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

				Expect(executable.ExecuteCall.Receives.Execution).To(Equal(packit.Execution{
					Args: []string{"build", "--tag", "some-tag", "."},
				}))
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

				Expect(executable.ExecuteCall.Receives.Execution).To(Equal(packit.Execution{
					Args: []string{"build", "--file", "some-file", "."},
				}))
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

				Expect(executable.ExecuteCall.Receives.Execution).To(Equal(packit.Execution{
					Args: []string{"build", "some-context"},
					Dir:  "some-context",
				}))
			})
		})

		Context("failure cases", func() {
			Context("when the executable cannot execute", func() {
				BeforeEach(func() {
					executable.ExecuteCall.Returns.Err = errors.New("failed to execute")
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

			Expect(executable.ExecuteCall.Receives.Execution).To(Equal(packit.Execution{
				Args: []string{"run", "some-image"},
			}))
		})

		Context("when given the network option to specify the network the container should use", func() {
			It("executes the run command with the --network flag", func() {
				stdout, stderr, err := cli.Run("some-image", docker.RunOptions{
					Network: "some-network",
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(stdout).To(Equal("some-stdout-output"))
				Expect(stderr).To(Equal("some-stderr-output"))
				Expect(executable.ExecuteCall.Receives.Execution).To(Equal(packit.Execution{
					Args: []string{"run", "--network", "some-network", "some-image"},
				}))
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
				Expect(executable.ExecuteCall.Receives.Execution).To(Equal(packit.Execution{
					Args: []string{"run", "--rm", "some-image"},
				}))
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
				Expect(executable.ExecuteCall.Receives.Execution).To(Equal(packit.Execution{
					Args: []string{"run", "--tty", "some-image"},
				}))
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
				Expect(executable.ExecuteCall.Receives.Execution).To(Equal(packit.Execution{
					Args: []string{"run", "some-image", "bash", "-c", "some-command"},
				}))
			})
		})

		Context("failure cases", func() {
			Context("when the executable fails to execute", func() {
				BeforeEach(func() {
					executable.ExecuteCall.Returns.Err = errors.New("failed to execute")
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

			Expect(executable.ExecuteCall.Receives.Execution).To(Equal(packit.Execution{
				Args: []string{"image", "rm", "some-image"},
			}))
		})

		Context("when given the force option", func() {
			It("executes the image rm command with the --force flag", func() {
				stdout, stderr, err := cli.RemoveImage("some-image", docker.RemoveImageOptions{
					Force: true,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(stdout).To(Equal("some-stdout-output"))
				Expect(stderr).To(Equal("some-stderr-output"))

				Expect(executable.ExecuteCall.Receives.Execution).To(Equal(packit.Execution{
					Args: []string{"image", "rm", "--force", "some-image"},
				}))
			})
		})

		Context("failure cases", func() {
			Context("when the executable fails to execute", func() {
				BeforeEach(func() {
					executable.ExecuteCall.Returns.Err = errors.New("failed to execute")
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
