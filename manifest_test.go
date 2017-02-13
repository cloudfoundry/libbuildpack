package libbuildpack_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"

	bp "github.com/cloudfoundry/libbuildpack"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/jarcoal/httpmock.v1"
)

var _ = Describe("Manifest", func() {
	var (
		manifest    bp.Manifest
		manifestDir string
		err         error
	)

	BeforeEach(func() {
		manifestDir = "fixtures/manifest/standard"
		httpmock.Reset()
	})
	JustBeforeEach(func() {
		manifest, err = bp.NewManifest(manifestDir)
		Expect(err).To(BeNil())
	})

	Describe("NewManifest", func() {
		It("has a language", func() {
			Expect(manifest.Language()).To(Equal("dotnet-core"))
		})
	})

	Describe("FetchDependency", func() {
		var tmpdir, outputFile string

		BeforeEach(func() {
			manifestDir = "fixtures/manifest/fetch"
			tmpdir, err = ioutil.TempDir("", "downloads")
			Expect(err).To(BeNil())
			outputFile = filepath.Join(tmpdir, "out.tgz")
		})
		AfterEach(func() { err = os.RemoveAll(tmpdir); Expect(err).To(BeNil()) })

		Context("uncached", func() {
			Context("url exists and matches md5", func() {
				BeforeEach(func() {
					httpmock.RegisterResponder("GET", "https://example.com/dependencies/thing-1-linux-x64.tgz",
						httpmock.NewStringResponder(200, "exciting binary data"))
				})

				It("downloads the file to the requested location", func() {
					err = manifest.FetchDependency(bp.Dependency{Name: "thing", Version: "1"}, outputFile)

					Expect(err).To(BeNil())
					Expect(ioutil.ReadFile(outputFile)).To(Equal([]byte("exciting binary data")))
				})

				It("makes intermediate directories", func() {
					outputFile = filepath.Join(tmpdir, "notexist", "out.tgz")
					err = manifest.FetchDependency(bp.Dependency{Name: "thing", Version: "1"}, outputFile)

					Expect(err).To(BeNil())
					Expect(ioutil.ReadFile(outputFile)).To(Equal([]byte("exciting binary data")))
				})
			})

			Context("url returns 404", func() {
				BeforeEach(func() {
					httpmock.RegisterResponder("GET", "https://example.com/dependencies/thing-1-linux-x64.tgz",
						httpmock.NewStringResponder(404, "exciting binary data"))
				})
				It("raises error", func() {
					err = manifest.FetchDependency(bp.Dependency{Name: "thing", Version: "1"}, outputFile)

					Expect(err).ToNot(BeNil())
				})

				It("alerts the user that the url could not be downloaded", func() {
					buf := new(bytes.Buffer)
					bp.Log.SetOutput(buf)

					err = manifest.FetchDependency(bp.Dependency{Name: "thing", Version: "1"}, outputFile)

					Expect(buf.String()).To(ContainSubstring("**ERROR** Could not download: 404"))
					Expect(buf.String()).ToNot(ContainSubstring("to ["))
				})

				It("outputfile does not exist", func() {
					err = manifest.FetchDependency(bp.Dependency{Name: "thing", Version: "1"}, outputFile)

					Expect(outputFile).ToNot(BeAnExistingFile())
				})
			})

			Context("url exists but does not match md5", func() {
				BeforeEach(func() {
					httpmock.RegisterResponder("GET", "https://example.com/dependencies/thing-1-linux-x64.tgz",
						httpmock.NewStringResponder(200, "other data"))
				})
				It("raises error", func() {
					err = manifest.FetchDependency(bp.Dependency{Name: "thing", Version: "1"}, outputFile)

					Expect(err).ToNot(BeNil())
				})
				It("outputfile does not exist", func() {
					err = manifest.FetchDependency(bp.Dependency{Name: "thing", Version: "1"}, outputFile)

					Expect(outputFile).ToNot(BeAnExistingFile())
				})
			})

		})

		Context("cached", func() {
			var dependenciesDir string

			BeforeEach(func() {
				var err error
				manifestDir, err = ioutil.TempDir("", "cached")
				Expect(err).To(BeNil())

				dependenciesDir = filepath.Join(manifestDir, "dependencies")
				os.MkdirAll(dependenciesDir, 0755)

				data, err := ioutil.ReadFile("fixtures/manifest/fetch/manifest.yml")
				Expect(err).To(BeNil())

				err = ioutil.WriteFile(filepath.Join(manifestDir, "manifest.yml"), data, 0644)
				Expect(err).To(BeNil())

				outputFile = filepath.Join(tmpdir, "out.tgz")
			})

			Context("url exists cached on disk and matches md5", func() {
				BeforeEach(func() {
					ioutil.WriteFile(filepath.Join(dependenciesDir, "https___example.com_dependencies_thing-2-linux-x64.tgz"), []byte("awesome binary data"), 0644)
				})
				It("copies the cached file to outputFile", func() {
					err = manifest.FetchDependency(bp.Dependency{Name: "thing", Version: "2"}, outputFile)

					Expect(err).To(BeNil())
					Expect(ioutil.ReadFile(outputFile)).To(Equal([]byte("awesome binary data")))
				})
				It("makes intermediate directories", func() {
					outputFile = filepath.Join(tmpdir, "notexist", "out.tgz")
					err = manifest.FetchDependency(bp.Dependency{Name: "thing", Version: "2"}, outputFile)

					Expect(err).To(BeNil())
					Expect(ioutil.ReadFile(outputFile)).To(Equal([]byte("awesome binary data")))
				})
			})

			Context("url exists cached on disk and does not match md5", func() {
				BeforeEach(func() {
					ioutil.WriteFile(filepath.Join(dependenciesDir, "https___example.com_dependencies_thing-2-linux-x64.tgz"), []byte("different binary data"), 0644)
				})
				It("raises error", func() {
					err = manifest.FetchDependency(bp.Dependency{Name: "thing", Version: "2"}, outputFile)

					Expect(err).ToNot(BeNil())
				})
				It("outputfile does not exist", func() {
					err = manifest.FetchDependency(bp.Dependency{Name: "thing", Version: "2"}, outputFile)

					Expect(outputFile).ToNot(BeAnExistingFile())
				})
			})

			Context("url is not cached on disk", func() {
				It("raises error", func() {
					err = manifest.FetchDependency(bp.Dependency{Name: "thing", Version: "2"}, outputFile)

					Expect(err).ToNot(BeNil())
				})
			})
		})
	})

	Describe("DefaultVersion", func() {
		Context("requested name exists (once)", func() {
			It("returns the default", func() {
				dep, err := manifest.DefaultVersion("node")
				Expect(err).To(BeNil())

				Expect(dep).To(Equal(bp.Dependency{Name: "node", Version: "6.9.4"}))
			})
		})

		Context("requested name exists (twice)", func() {
			BeforeEach(func() { manifestDir = "fixtures/manifest/duplicate" })
			It("returns an error", func() {
				_, err := manifest.DefaultVersion("bower")
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("found 2 default versions for bower"))
			})
		})
		Context("requested name does not exist", func() {
			It("returns an error", func() {
				_, err := manifest.DefaultVersion("notexist")
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("no default version for notexist"))
			})
		})
	})
})
