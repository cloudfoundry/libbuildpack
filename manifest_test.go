package libbuildpack_test

import (
	"bytes"
	"errors"
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
		version     string
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

	Describe("CheckStackSupport", func() {
		var (
			oldCfStack string
		)

		BeforeEach(func() { oldCfStack = os.Getenv("CF_STACK") })
		AfterEach(func() { err = os.Setenv("CF_STACK", oldCfStack); Expect(err).To(BeNil()) })

		Context("Stack is supported", func() {
			BeforeEach(func() {
				manifestDir = "fixtures/manifest/stacks"
				err = os.Setenv("CF_STACK", "cflinuxfs2")
				Expect(err).To(BeNil())
			})

			It("returns nil", func() {
				Expect(manifest.CheckStackSupport()).To(Succeed())
			})

			Context("by a single dependency", func() {
				BeforeEach(func() {
					manifestDir = "fixtures/manifest/stacks"
					err = os.Setenv("CF_STACK", "xenial")
					Expect(err).To(BeNil())
				})
				It("returns nil", func() {
					Expect(manifest.CheckStackSupport()).To(Succeed())
				})
			})
		})

		Context("Stack is not supported", func() {
			BeforeEach(func() {
				err = os.Setenv("CF_STACK", "notastack")
				Expect(err).To(BeNil())
			})

			It("returns nil", func() {
				Expect(manifest.CheckStackSupport()).To(MatchError(errors.New("required stack notastack was not found")))
			})
		})
	})

	Describe("Version", func() {
		Context("VERSION file exists", func() {
			It("returns the version", func() {
				version, err = manifest.Version()
				Expect(err).To(BeNil())

				Expect(version).To(Equal("99.99"))
			})
		})

		Context("VERSION file does not exist", func() {
			BeforeEach(func() {
				manifestDir = "fixtures/manifest/duplicate"
			})

			It("returns an error", func() {
				version, err = manifest.Version()
				Expect(version).To(Equal(""))
				Expect(err).ToNot(BeNil())

				Expect(err.Error()).To(ContainSubstring("unable to read VERSION file"))
			})
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

	Describe("CheckBuildpackVersion", func() {
		var (
			cacheDir string
			buffer   *bytes.Buffer
		)
		BeforeEach(func() {
			cacheDir, err = ioutil.TempDir("", "cache")

			buffer = new(bytes.Buffer)
			bp.Log.SetOutput(buffer)
		})
		AfterEach(func() {
			err = os.RemoveAll(cacheDir)
			Expect(err).To(BeNil())

			bp.Log.SetOutput(ioutil.Discard)
		})

		Context("BUILDPACK_METADATA exists", func() {
			Context("The language does not match", func() {
				BeforeEach(func() {
					metadata := "---\nlanguage: diffLang\nversion: 99.99"
					ioutil.WriteFile(filepath.Join(cacheDir, "BUILDPACK_METADATA"), []byte(metadata), 0666)
				})

				It("Does not log anything", func() {
					manifest.CheckBuildpackVersion(cacheDir)
					Expect(buffer.String()).To(Equal(""))
				})
			})
			Context("The language matches", func() {
				Context("The version matches", func() {
					BeforeEach(func() {
						metadata := "---\nlanguage: dotnet-core\nversion: 99.99"
						ioutil.WriteFile(filepath.Join(cacheDir, "BUILDPACK_METADATA"), []byte(metadata), 0666)
					})

					It("Does not log anything", func() {
						manifest.CheckBuildpackVersion(cacheDir)
						Expect(buffer.String()).To(Equal(""))

					})
				})

				Context("The version does not match", func() {
					BeforeEach(func() {
						metadata := "---\nlanguage: dotnet-core\nversion: 33.99"
						ioutil.WriteFile(filepath.Join(cacheDir, "BUILDPACK_METADATA"), []byte(metadata), 0666)
					})

					It("Logs a warning that the buildpack version has changed", func() {
						manifest.CheckBuildpackVersion(cacheDir)
						Expect(buffer.String()).To(ContainSubstring("buildpack version changed from 33.99 to 99.99"))

					})
				})
			})
		})

		Context("BUILDPACK_METADATA does not exist", func() {
			It("Does not log anything", func() {
				manifest.CheckBuildpackVersion(cacheDir)
				Expect(buffer.String()).To(Equal(""))

			})
		})
	})
})
