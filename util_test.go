package libbuildpack_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	bp "github.com/cloudfoundry/libbuildpack"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Util", func() {
	Describe("Unzip", func() {
		var (
			tmpdir string
			err    error
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
			tmpdir string
			err    error
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
})
