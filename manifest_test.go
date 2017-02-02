package buildpack_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	be "github.com/sesmith177/buildpack-extensions"
	"gopkg.in/jarcoal/httpmock.v1"
)

var _ = Describe("Manifest", func() {
	var (
		manifest     *be.Manifest
		manifestFile string
		err          error
	)

	BeforeSuite(func() { httpmock.Activate() })
	AfterSuite(func() { httpmock.DeactivateAndReset() })

	BeforeEach(func() {
		manifestFile = "fixtures/manifest.yml"
		httpmock.Reset()
	})
	JustBeforeEach(func() {
		manifest, err = be.NewManifest(manifestFile)
		Expect(err).To(BeNil())
	})

	Describe("NewManifest", func() {
		It("has a language", func() {
			Expect(manifest.Language).To(Equal("dotnet-core"))
		})
	})

	Describe("FetchDependency", func() {
		var tmpdir string

		BeforeEach(func() {
			manifestFile = "fixtures/manifest_fetch.yml"
			tmpdir, err = ioutil.TempDir("", "downloads")
			Expect(err).To(BeNil())
		})
		AfterEach(func() { err = os.RemoveAll(tmpdir); Expect(err).To(BeNil()) })

		Context("uncached", func() {
			Context("url exists and matches md5", func() {
				BeforeEach(func() {
					httpmock.RegisterResponder("GET", "https://example.com/dependencies/file.tgz",
						httpmock.NewStringResponder(200, "exciting binary data"))
				})

				It("downloads the file to the requested location", func() {
					outputFile := filepath.Join(tmpdir, "out.tgz")
					err = manifest.FetchDependency(be.Dependency{Name: "thing", Version: "1"}, outputFile)

					Expect(err).To(BeNil())
					Expect(ioutil.ReadFile(outputFile)).To(Equal([]byte("exciting binary data")))
				})

				It("makes intermediate directories", func() {
					outputFile := filepath.Join(tmpdir, "notexist", "out.tgz")
					err = manifest.FetchDependency(be.Dependency{Name: "thing", Version: "1"}, outputFile)

					Expect(err).To(BeNil())
					Expect(ioutil.ReadFile(outputFile)).To(Equal([]byte("exciting binary data")))
				})
			})

			Context("url exists but does not match md5", func() {
				BeforeEach(func() {
					httpmock.RegisterResponder("GET", "https://example.com/dependencies/file.tgz",
						httpmock.NewStringResponder(200, "other data"))
				})
				It("raises error", func() {
					outputFile := filepath.Join(tmpdir, "out.tgz")
					err = manifest.FetchDependency(be.Dependency{Name: "thing", Version: "1"}, outputFile)

					Expect(err).ToNot(BeNil())
				})
				It("outputfile does not exist", func() {
					outputFile := filepath.Join(tmpdir, "out.tgz")
					err = manifest.FetchDependency(be.Dependency{Name: "thing", Version: "1"}, outputFile)

					Expect(outputFile).ToNot(BeAnExistingFile())
				})
			})

		})

		PContext("cached", func() {})
	})

	Describe("DefaultVersion", func() {
		Context("requested name exists (once)", func() {
			It("returns the default", func() {
				Expect(manifest.DefaultVersion("node")).To(Equal("6.9.4"))
			})
		})

		Context("requested name exists (twice)", func() {
			BeforeEach(func() { manifestFile = "fixtures/manifest_duplicate_default.yml" })
			It("returns an buildpack error", func() {
				_, err := manifest.DefaultVersion("bower")
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("found 2 default versions for bower"))
				Expect(err.(be.Error).BuildpackError()).To(ContainSubstring("misconfigured for 'default_versions'"))
			})
		})
		Context("requested name does not exist", func() {
			It("returns an buildpack error", func() {
				_, err := manifest.DefaultVersion("notexist")
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("no default version for notexist"))
				Expect(err.(be.Error).BuildpackError()).To(ContainSubstring("misconfigured for 'default_versions'"))
			})
		})
	})
})
