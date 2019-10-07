package docker_test

import (
	"bytes"
	"io"
	"os"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry/libbuildpack/cutlass/docker"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("BuildStagingDockerfile", func() {
	var logger lager.Logger

	BeforeEach(func() {
		logger = lager.NewLogger("cutlass")
	})

	It("returns a dockerfile used for staging", func() {
		dockerfile := docker.BuildStagingDockerfile(logger, "/some/fixture-path", "/some/buildpack-path", []string{"some-env", "other-env"})
		Expect(dockerfile.String()).To(MatchLines(`
FROM cloudfoundry/cflinuxfs3
ENV CF_STACK cflinuxfs3
ENV VCAP_APPLICATION {}
ENV some-env
ENV other-env
ADD /some/fixture-path /tmp/staged/
ADD /some/buildpack-path /tmp/
RUN mkdir -p /buildpack/0
RUN mkdir -p /tmp/cache
RUN unzip /tmp/buildpack-path -d /buildpack
RUN mv /usr/sbin/tcpdump /usr/bin/tcpdump`))
	})

	Context("when the CF_STACK environment variable is set", func() {
		BeforeEach(func() {
			Expect(os.Setenv("CF_STACK", "some-stack")).To(Succeed())
		})

		AfterEach(func() {
			Expect(os.Unsetenv("CF_STACK")).To(Succeed())
		})

		It("returns a dockerfile used for staging", func() {
			dockerfile := docker.BuildStagingDockerfile(logger, "/some/fixture-path", "/some/buildpack-path", []string{"some-env", "other-env"})
			Expect(dockerfile.String()).To(MatchLines(`
FROM cloudfoundry/some-stack
ENV CF_STACK some-stack
ENV VCAP_APPLICATION {}
ENV some-env
ENV other-env
ADD /some/fixture-path /tmp/staged/
ADD /some/buildpack-path /tmp/
RUN mkdir -p /buildpack/0
RUN mkdir -p /tmp/cache
RUN unzip /tmp/buildpack-path -d /buildpack
RUN mv /usr/sbin/tcpdump /usr/bin/tcpdump`))
		})
	})

	Context("when the CF_STACK_DOCKER_IMAGE environment variable is set", func() {
		BeforeEach(func() {
			Expect(os.Setenv("CF_STACK_DOCKER_IMAGE", "some-stack-docker-image")).To(Succeed())
		})

		AfterEach(func() {
			Expect(os.Unsetenv("CF_STACK_DOCKER_IMAGE")).To(Succeed())
		})

		It("returns a dockerfile used for staging", func() {
			dockerfile := docker.BuildStagingDockerfile(logger, "/some/fixture-path", "/some/buildpack-path", []string{"some-env", "other-env"})
			Expect(dockerfile.String()).To(MatchLines(`
FROM some-stack-docker-image
ENV CF_STACK cflinuxfs3
ENV VCAP_APPLICATION {}
ENV some-env
ENV other-env
ADD /some/fixture-path /tmp/staged/
ADD /some/buildpack-path /tmp/
RUN mkdir -p /buildpack/0
RUN mkdir -p /tmp/cache
RUN unzip /tmp/buildpack-path -d /buildpack
RUN mv /usr/sbin/tcpdump /usr/bin/tcpdump`))
		})
	})
})

var _ = Describe("DockerfileInstruction", func() {
	Describe("String", func() {
		DescribeTable("covering all types of instruction",
			func(instructionType docker.DockerfileInstructionType, output string) {
				instruction := docker.DockerfileInstruction{
					Type:    instructionType,
					Content: "some-content",
				}
				Expect(instruction.String()).To(Equal(output))
			},

			Entry("FROM", docker.DockerfileInstructionTypeFROM, "FROM some-content"),
			Entry("ADD", docker.DockerfileInstructionTypeADD, "ADD some-content"),
			Entry("RUN", docker.DockerfileInstructionTypeRUN, "RUN some-content"),
			Entry("ENV", docker.DockerfileInstructionTypeENV, "ENV some-content"),
		)
	})
})

var _ = Describe("Dockerfile", func() {
	Describe("String", func() {
		It("returns a string representation of the Dockerfile", func() {
			dockerfile := docker.NewDockerfile("some-base-image", docker.DockerfileInstruction{
				Type:    docker.DockerfileInstructionTypeRUN,
				Content: "some-content",
			})
			content := dockerfile.String()
			Expect(content).To(MatchLines(`
FROM some-base-image
RUN some-content`))
		})
	})

	Describe("Read", func() {
		It("reads the contents of the dockerfile into the given byte slice", func() {
			buffer := make([]byte, 40) // NOTE: this must be longer than the Dockerfile content

			dockerfile := docker.NewDockerfile("some-base-image", docker.DockerfileInstruction{
				Type:    docker.DockerfileInstructionTypeRUN,
				Content: "some-content",
			})

			count, err := dockerfile.Read(buffer)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(buffer)).To(MatchLines(`
FROM some-base-image
RUN some-content`))
			Expect(count).To(Equal(38))
		})

		It("returns io.EOF when there is nothing more to read", func() {
			buffer := bytes.NewBuffer([]byte{})

			dockerfile := docker.NewDockerfile("some-base-image", docker.DockerfileInstruction{
				Type:    docker.DockerfileInstructionTypeRUN,
				Content: "some-content",
			})

			count, err := io.Copy(buffer, dockerfile)
			Expect(err).NotTo(HaveOccurred())
			Expect(buffer.String()).To(MatchLines(`
FROM some-base-image
RUN some-content`))
			Expect(count).To(Equal(int64(38)))
		})
	})
})
