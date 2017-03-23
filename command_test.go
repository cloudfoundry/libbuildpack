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
		cmd    bp.Command
	)

	JustBeforeEach(func() {
		buffer = new(bytes.Buffer)

		cmd = bp.NewCommand(exe, args...)
		cmd.SetOutput(buffer)
	})

	Context("valid command", func() {
		BeforeEach(func() {
			exe = "ls"
			args = []string{"-l", "fixtures"}

		})

		It("runs the command with the output in the right location", func() {
			err := cmd.Run()
			Expect(err).To(BeNil())

			Expect(buffer.String()).To(ContainSubstring("thing.tgz"))
		})
	})

	Context("invalid command", func() {
		BeforeEach(func() {
			exe = "ls"
			args = []string{"-l", "not/a/dir"}

		})

		It("runs the command and returns an eror", func() {
			err := cmd.Run()
			Expect(err).NotTo(BeNil())

			Expect(buffer.String()).To(Equal("ls: not/a/dir: No such file or directory\n"))
		})
	})
})
