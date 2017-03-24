package libbuildpack_test

import (
	"bytes"

	bp "github.com/cloudfoundry/libbuildpack"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Command", func() {
	var (
		buffer *bytes.Buffer
		exe    string
		args   []string
		cmd    bp.CommandRunner
	)

	JustBeforeEach(func() {
		buffer = new(bytes.Buffer)

		cmd = bp.NewCommandRunner()
		cmd.SetOutput(buffer)
	})

	AfterEach(func() {
		cmd.Reset()
	})

	Context("valid command", func() {
		BeforeEach(func() {
			exe = "ls"
			args = []string{"-l", "fixtures"}
		})

		It("runs the command with the output in the right location", func() {
			err := cmd.Run(exe, args...)
			Expect(err).To(BeNil())

			Expect(buffer.String()).To(ContainSubstring("thing.tgz"))
		})
	})
	Context("changing directory", func() {
		BeforeEach(func() {
			exe = "pwd"
			args = []string{}
		})

		It("runs the command with the output in the right location", func() {
			cmd.SetDir("fixtures")
			err := cmd.Run(exe, args...)
			Expect(err).To(BeNil())

			Expect(buffer.String()).To(ContainSubstring("libbuildpack/fixtures"))
		})
	})

	Context("capturing output", func() {
		BeforeEach(func() {
			exe = "ls"
			args = []string{"-l", "fixtures"}
		})

		It("returns the output as a string", func() {
			output, err := cmd.CaptureOutput(exe, args...)
			Expect(err).To(BeNil())

			Expect(output).To(ContainSubstring("thing.tgz"))
		})
	})

	Context("invalid command", func() {
		BeforeEach(func() {
			exe = "ls"
			args = []string{"-l", "not/a/dir"}

		})

		It("runs the command and returns an eror", func() {
			err := cmd.Run(exe, args...)
			Expect(err).NotTo(BeNil())

			Expect(buffer.String()).To(ContainSubstring("No such file or directory"))
		})
	})
})
