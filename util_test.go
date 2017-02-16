package libbuildpack_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"

	bp "github.com/cloudfoundry/libbuildpack"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Util", func() {
	Describe("Unzip", func() {
		var (
			tmpdir   string
			err      error
			fileInfo os.FileInfo
		)
		BeforeEach(func() {
			tmpdir, err = ioutil.TempDir("", "exploded")
			Expect(err).To(BeNil())
		})
		AfterEach(func() { err = os.RemoveAll(tmpdir); Expect(err).To(BeNil()) })

		Context("with a valid zip file", func() {
			It("extracts a file at the root", func() {
				err = bp.ExtractZip("fixtures/thing.zip", tmpdir)
				Expect(err).To(BeNil())

				Expect(filepath.Join(tmpdir, "root.txt")).To(BeAnExistingFile())
				Expect(ioutil.ReadFile(filepath.Join(tmpdir, "root.txt"))).To(Equal([]byte("root\n")))
			})

			It("extracts a nested file", func() {
				err = bp.ExtractZip("fixtures/thing.zip", tmpdir)
				Expect(err).To(BeNil())

				Expect(filepath.Join(tmpdir, "thing", "bin", "file2.exe")).To(BeAnExistingFile())
				Expect(ioutil.ReadFile(filepath.Join(tmpdir, "thing", "bin", "file2.exe"))).To(Equal([]byte("progam2\n")))
			})

			It("preserves the file mode", func() {
				err = bp.ExtractZip("fixtures/thing.zip", tmpdir)
				Expect(err).To(BeNil())

				Expect(filepath.Join(tmpdir, "thing", "bin", "file2.exe")).To(BeAnExistingFile())
				fileInfo, err = os.Stat(filepath.Join(tmpdir, "thing", "bin", "file2.exe"))

				Expect(fileInfo.Mode()).To(Equal(os.FileMode(0755)))
			})
		})

		Context("with a missing zip file", func() {
			It("returns an error", func() {
				err = bp.ExtractZip("fixtures/notexist.zip", tmpdir)
				Expect(err).ToNot(BeNil())
			})
		})

		Context("with an invalid zip file", func() {
			It("returns an error", func() {
				err = bp.ExtractZip("fixtures/manifest.yml", tmpdir)
				Expect(err).ToNot(BeNil())
			})
		})
	})

	Describe("Untar", func() {
		var (
			tmpdir   string
			err      error
			fileInfo os.FileInfo
		)
		BeforeEach(func() {
			tmpdir, err = ioutil.TempDir("", "exploded")
			Expect(err).To(BeNil())
		})
		AfterEach(func() { err = os.RemoveAll(tmpdir); Expect(err).To(BeNil()) })

		Context("with a valid tar file", func() {
			It("extracts a file at the root", func() {
				err = bp.ExtractTarGz("fixtures/thing.tgz", tmpdir)
				Expect(err).To(BeNil())

				Expect(filepath.Join(tmpdir, "root.txt")).To(BeAnExistingFile())
				Expect(ioutil.ReadFile(filepath.Join(tmpdir, "root.txt"))).To(Equal([]byte("root\n")))
			})
			It("extracts a nested file", func() {
				err = bp.ExtractTarGz("fixtures/thing.tgz", tmpdir)
				Expect(err).To(BeNil())

				Expect(filepath.Join(tmpdir, "thing", "bin", "file2.exe")).To(BeAnExistingFile())
				Expect(ioutil.ReadFile(filepath.Join(tmpdir, "thing", "bin", "file2.exe"))).To(Equal([]byte("progam2\n")))
			})
			It("preserves the file mode", func() {
				err = bp.ExtractTarGz("fixtures/thing.tgz", tmpdir)
				Expect(err).To(BeNil())

				Expect(filepath.Join(tmpdir, "thing", "bin", "file2.exe")).To(BeAnExistingFile())
				fileInfo, err = os.Stat(filepath.Join(tmpdir, "thing", "bin", "file2.exe"))

				Expect(fileInfo.Mode()).To(Equal(os.FileMode(0755)))
			})
		})

		Context("with a missing tar file", func() {
			It("returns an error", func() {
				err = bp.ExtractTarGz("fixtures/notexist.tgz", tmpdir)
				Expect(err).ToNot(BeNil())
			})
		})

		Context("with an invalid tar file", func() {
			It("returns an error", func() {
				err = bp.ExtractTarGz("fixtures/manifest.yml", tmpdir)
				Expect(err).ToNot(BeNil())
			})
		})
	})

	Describe("CopyFile", func() {
		var (
			tmpdir   string
			err      error
			fileInfo os.FileInfo
			oldMode  os.FileMode
			oldUmask int
		)
		BeforeEach(func() {
			var fh *os.File
			sourceFile := "fixtures/source.txt"

			tmpdir, err = ioutil.TempDir("", "copy")
			Expect(err).To(BeNil())

			fileInfo, err = os.Stat(sourceFile)
			Expect(err).To(BeNil())
			oldMode = fileInfo.Mode()

			fh, err = os.Open(sourceFile)
			Expect(err).To(BeNil())
			defer fh.Close()

			err = fh.Chmod(0742)
			Expect(err).To(BeNil())

			oldUmask = syscall.Umask(0000)

		})
		AfterEach(func() {
			var fh *os.File
			sourceFile := "fixtures/source.txt"

			fh, err = os.Open(sourceFile)
			Expect(err).To(BeNil())
			defer fh.Close()

			err = fh.Chmod(oldMode)
			Expect(err).To(BeNil())

			err = os.RemoveAll(tmpdir)
			Expect(err).To(BeNil())

			syscall.Umask(oldUmask)
		})

		Context("with a valid source file", func() {
			It("copies the file", func() {
				err = bp.CopyFile("fixtures/source.txt", filepath.Join(tmpdir, "out.txt"))
				Expect(err).To(BeNil())

				Expect(filepath.Join(tmpdir, "out.txt")).To(BeAnExistingFile())
				Expect(ioutil.ReadFile(filepath.Join(tmpdir, "out.txt"))).To(Equal([]byte("a file\n")))
			})

			It("preserves the file mode", func() {
				err = bp.CopyFile("fixtures/source.txt", filepath.Join(tmpdir, "out.txt"))
				Expect(err).To(BeNil())

				Expect(filepath.Join(tmpdir, "out.txt")).To(BeAnExistingFile())
				fileInfo, err = os.Stat(filepath.Join(tmpdir, "out.txt"))

				Expect(fileInfo.Mode()).To(Equal(os.FileMode(0742)))
			})
		})

		Context("with a missing file", func() {
			It("returns an error", func() {
				err = bp.ExtractTarGz("fixtures/notexist.txt", tmpdir)
				Expect(err).ToNot(BeNil())
			})
		})
	})
})
