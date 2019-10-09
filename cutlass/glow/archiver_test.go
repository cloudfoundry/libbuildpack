package glow_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack/cutlass/glow"
	"github.com/cloudfoundry/libbuildpack/cutlass/glow/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Archiver", func() {
	var (
		tmpDir string

		packager *fakes.Packager
		archiver glow.Archiver
	)

	BeforeEach(func() {
		var err error
		tmpDir, err = ioutil.TempDir("", "archiver")
		Expect(err).NotTo(HaveOccurred())

		packager = &fakes.Packager{}
		packager.PackageCall.Returns.Stderr = "Packaged Shimmed Buildpack at: some-buildpack-file.zip"

		archiver = glow.NewArchiver(packager)
	})

	AfterEach(func() {
		Expect(os.RemoveAll(tmpDir)).To(Succeed())
	})

	Describe("Archive", func() {
		BeforeEach(func() {
			err := ioutil.WriteFile(filepath.Join(tmpDir, "VERSION"), []byte("1.2.3"), 0644)
			Expect(err).NotTo(HaveOccurred())
		})

		It("creates an archive of the given buildpack", func() {
			path, err := archiver.Archive(tmpDir, "some-stack", "some-tag", true)
			Expect(err).NotTo(HaveOccurred())
			Expect(path).To(Equal(filepath.Join(tmpDir, "some-buildpack-file.zip")))

			Expect(packager.PackageCall.Receives.Dir).To(Equal(tmpDir))
			Expect(packager.PackageCall.Receives.Stack).To(Equal("some-stack"))
			Expect(packager.PackageCall.Receives.Options).To(Equal(glow.PackageOptions{
				Cached:  true,
				Version: "1.2.3-some-tag",
			}))
		})

		Context("failure cases", func() {
			Context("when the VERSION file is missing", func() {
				BeforeEach(func() {
					Expect(os.Remove(filepath.Join(tmpDir, "VERSION"))).To(Succeed())
				})

				It("returns an error", func() {
					_, err := archiver.Archive(tmpDir, "some-stack", "some-tag", true)
					Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
				})
			})

			Context("when the packager returns an error", func() {
				BeforeEach(func() {
					packager.PackageCall.Returns.Err = errors.New("failed to package")
				})

				It("returns an error", func() {
					_, err := archiver.Archive(tmpDir, "some-stack", "some-tag", true)
					Expect(err).To(MatchError("running package command failed: failed to package"))
				})
			})

			Context("when the output does not contain the archive file path", func() {
				BeforeEach(func() {
					packager.PackageCall.Returns.Stderr = "No archive in this output"
				})

				It("returns an error", func() {
					_, err := archiver.Archive(tmpDir, "some-stack", "some-tag", true)
					Expect(err).To(MatchError("failed to find archive file path in output:\nNo archive in this output"))
				})
			})
		})
	})
})
