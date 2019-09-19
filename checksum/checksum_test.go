package checksum_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/cloudfoundry/libbuildpack/checksum"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Checksum", func() {
	var (
		dir   string
		exec  func() error
		lines []string
	)

	BeforeEach(func() {
		var err error
		dir, err = ioutil.TempDir("", "checksum")
		Expect(err).To(BeNil())

		Expect(os.MkdirAll(filepath.Join(dir, ".cloudfoundry"), 0755)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(dir, "a", "b"), 0755)).To(Succeed())
		Expect(ioutil.WriteFile(filepath.Join(dir, "a/b", "file"), []byte("hi"), 0644)).To(Succeed())

		lines = []string{}
		exec = func() error { return nil }
	})

	AfterEach(func() {
		Expect(os.RemoveAll(dir)).To(Succeed())
	})

	debug := func(format string, args ...interface{}) {
		lines = append(lines, fmt.Sprintf(format, args...))
	}

	Describe("Do", func() {
		Context("when the directory is unchanged", func() {
			It("reports the current directory checksum", func() {
				Expect(checksum.Do(dir, debug, exec)).To(Succeed())
				Expect(lines).To(Equal([]string{
					"Checksum Before (" + dir + "): 3e673106d28d587c5c01b3582bf15a50",
					"Checksum After (" + dir + "): 3e673106d28d587c5c01b3582bf15a50",
				}))
			})
		})

		Context("when a file is changed", func() {
			BeforeEach(func() {
				exec = func() error {
					time.Sleep(10 * time.Millisecond)
					return ioutil.WriteFile(filepath.Join(dir, "a/b", "file"), []byte("bye"), 0644)
				}
			})

			It("reports the current directory checksum", func() {
				Expect(checksum.Do(dir, debug, exec)).To(Succeed())
				Expect(lines).To(Equal([]string{
					"Checksum Before (" + dir + "): 3e673106d28d587c5c01b3582bf15a50",
					"Checksum After (" + dir + "): e01956670269656ae69872c0672592ae",
					"Below files changed:",
					"./a/b/file",
				}))
			})
		})

		Context("when a file is added", func() {
			BeforeEach(func() {
				exec = func() error {
					time.Sleep(10 * time.Millisecond)
					return ioutil.WriteFile(filepath.Join(dir, "a", "file"), []byte("new file"), 0644)
				}
			})

			It("reports the current directory checksum", func() {
				Expect(checksum.Do(dir, debug, exec)).To(Succeed())
				Expect(lines).To(Equal([]string{
					"Checksum Before (" + dir + "): 3e673106d28d587c5c01b3582bf15a50",
					"Checksum After (" + dir + "): 9fc7505dc69734c5d40c38a35017e1dc",
					"Below files changed:",
					"./a",
					"./a/file",
				}))
			})
		})

		Context("when the modified file is in the .cloudfoundry directory", func() {
			BeforeEach(func() {
				exec = func() error {
					time.Sleep(10 * time.Millisecond)
					return ioutil.WriteFile(filepath.Join(dir, ".cloudfoundry", "file"), []byte("bye"), 0644)
				}
			})

			It("does not report a checksum change", func() {
				Expect(checksum.Do(dir, debug, exec)).To(Succeed())
				Expect(lines).To(Equal([]string{
					"Checksum Before (" + dir + "): 3e673106d28d587c5c01b3582bf15a50",
					"Checksum After (" + dir + "): 3e673106d28d587c5c01b3582bf15a50",
				}))
			})
		})

		Context("when exec returns an error", func() {
			BeforeEach(func() {
				exec = func() error {
					return errors.New("some error")
				}
			})

			It("returns an error", func() {
				Expect(checksum.Do(dir, debug, exec)).To(MatchError("some error"))
			})
		})
	})
})
