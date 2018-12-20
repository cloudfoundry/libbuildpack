package shims_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/cloudfoundry/libbuildpack"
	"github.com/cloudfoundry/libbuildpack/ansicleaner"
	httpmock "gopkg.in/jarcoal/httpmock.v1"

	"github.com/cloudfoundry/libbuildpack/shims"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

//go:generate mockgen -source=supplier.go --destination=mocks_shims_test.go --package=shims_test

var _ = Describe("Shims", func() {

	Describe("Supplier", func() {
		var (
			supplier        shims.Supplier
			mockCtrl        *gomock.Controller
			mockDetector    *MockDetector
			binDir          string
			v2BuildDir      string
			v2BuildpacksDir string
			cnbAppDir       string
			v3BuildpacksDir string
			depsDir         string
			depsIndex       string
			groupMetadata   string
			layersDir       string
			orderMetadata   string
			planMetadata    string
			tempDir         string
		)

		BeforeEach(func() {
			var err error

			mockCtrl = gomock.NewController(GinkgoT())
			mockDetector = NewMockDetector(mockCtrl)

			tempDir, err = ioutil.TempDir("", "tmp")
			Expect(err).NotTo(HaveOccurred())

			v2BuildDir = filepath.Join(tempDir, "build")
			Expect(os.MkdirAll(v2BuildDir, 0777)).To(Succeed())

			cnbAppDir = filepath.Join(tempDir, "cnb-app")

			binDir = filepath.Join(tempDir, "bin")
			Expect(os.MkdirAll(binDir, 0777)).To(Succeed())

			v3BuildpacksDir = filepath.Join(tempDir, "cnbs")
			Expect(os.MkdirAll(v3BuildpacksDir, 0777)).To(Succeed())

			depsDir = filepath.Join(tempDir, "deps")
			depsIndex = "0"

			layersDir = filepath.Join(tempDir, "layers")
			Expect(os.MkdirAll(filepath.Join(layersDir, "config"), 0777)).To(Succeed())

			v2BuildpacksDir = filepath.Join(tempDir, "buildpacks")

			groupMetadata = filepath.Join(tempDir, "group.toml")

			planMetadata = filepath.Join(tempDir, "plan.toml")
		})

		JustBeforeEach(func() {
			Expect(os.MkdirAll(filepath.Join(depsDir, depsIndex), 0777)).To(Succeed())

			Expect(os.MkdirAll(filepath.Join(v2BuildpacksDir, depsIndex), 0777)).To(Succeed())
			orderMetadata = filepath.Join(v2BuildpacksDir, depsIndex, "order.toml")
			Expect(ioutil.WriteFile(orderMetadata, []byte(""), 0666)).To(Succeed())

			supplier = shims.Supplier{
				Detector:        mockDetector,
				BinDir:          binDir,
				V2AppDir:        v2BuildDir,
				V2BuildpackDir:  filepath.Join(v2BuildpacksDir, depsIndex),
				V3AppDir:        cnbAppDir,
				V3BuildpacksDir: v3BuildpacksDir,
				V2DepsDir:       depsDir,
				DepsIndex:       depsIndex,
				V3LayersDir:     layersDir,
				OrderMetadata:   orderMetadata,
				GroupMetadata:   groupMetadata,
				PlanMetadata:    planMetadata,
			}
		})

		AfterEach(func() {
			mockCtrl.Finish()
			Expect(os.RemoveAll(tempDir)).To(Succeed())
		})

		Context("EnsureNoV2AfterV3", func() {
			It("does not return an error when there are no V2 buildpacks", func() {
				Expect(supplier.EnsureNoV2AfterV3()).To(Succeed())
			})

			Context("when there are V2 buildacks that have already run", func() {
				BeforeEach(func() {
					depsIndex = "1"
					Expect(os.MkdirAll(filepath.Join(v2BuildpacksDir, "0"), 0777)).To(Succeed())
					Expect(os.MkdirAll(filepath.Join(depsDir, "0"), 0777)).To(Succeed())
				})

				It("does not return an error", func() {
					Expect(supplier.EnsureNoV2AfterV3()).To(Succeed())
				})
			})

			Context("when there are V2 buildpacks that have not run yet", func() {
				BeforeEach(func() {
					depsIndex = "1"
					Expect(os.MkdirAll(filepath.Join(v2BuildpacksDir, "0"), 0777)).To(Succeed())
					Expect(ioutil.WriteFile(filepath.Join(v2BuildpacksDir, "0", "order.toml"), []byte(""), 0777)).To(Succeed())
					Expect(os.MkdirAll(filepath.Join(v2BuildpacksDir, "2"), 0777)).To(Succeed())
					Expect(os.MkdirAll(filepath.Join(depsDir, "2"), 0777)).To(Succeed())
				})

				It("returns an error", func() {
					Expect(supplier.EnsureNoV2AfterV3()).To(MatchError("Cannot follow a v3 buildpack by a v2 buildpack."))
				})
			})
		})

		Context("GetDetectorOutput", func() {
			It("runs detection when group or plan metadata does not exist", func() {
				mockDetector.
					EXPECT().
					Detect()
				Expect(supplier.GetDetectorOutput()).To(Succeed())
			})

			It("does NOT run detection when group and plan metadata exists", func() {
				Expect(ioutil.WriteFile(groupMetadata, []byte(""), 0666)).To(Succeed())
				Expect(ioutil.WriteFile(planMetadata, []byte(""), 0666)).To(Succeed())

				mockDetector.
					EXPECT().
					Detect().
					Times(0)
				Expect(supplier.GetDetectorOutput()).To(Succeed())
			})
		})

		Context("MoveV3Layers", func() {
			BeforeEach(func() {
				Expect(os.MkdirAll(filepath.Join(layersDir, "config"), 0777)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(layersDir, "config", "metadata.toml"), []byte(""), 0666)).To(Succeed())

				Expect(os.MkdirAll(filepath.Join(layersDir, "layer"), 0777)).To(Succeed())
				Expect(os.MkdirAll(filepath.Join(layersDir, "anotherLayer"), 0777)).To(Succeed())
			})

			It("moves the layers to deps dir and metadata to build dir", func() {
				Expect(supplier.MoveV3Layers()).To(Succeed())
				Expect(filepath.Join(v2BuildDir, ".cloudfoundry", "metadata.toml")).To(BeAnExistingFile())
				Expect(filepath.Join(depsDir, "layer")).To(BeAnExistingFile())
				Expect(filepath.Join(depsDir, "anotherLayer")).To(BeAnExistingFile())
			})

		})

		Context("AddV2SupplyBuildpacks", func() {
			var (
				createDirs, createFiles []string
			)

			BeforeEach(func() {
				depsIndex = "2"
				createDirs = []string{"bin", "lib"}
				createFiles = []string{"config.yml"}
				for _, dir := range createDirs {
					Expect(os.MkdirAll(filepath.Join(depsDir, "0", dir), 0777)).To(Succeed())
				}

				for _, file := range createFiles {
					Expect(ioutil.WriteFile(filepath.Join(depsDir, "0", file), []byte(file), 0666)).To(Succeed())
				}

				Expect(ioutil.WriteFile(groupMetadata, []byte(""), 0666)).To(Succeed())
				Expect(ioutil.WriteFile(planMetadata, []byte(""), 0666)).To(Succeed())
			})

			It("copies v2 layers and metadata where v3 lifecycle expects them for build and launch", func() {
				By("not failing if a layer has already been moved")
				Expect(supplier.AddV2SupplyBuildpacks()).To(Succeed())

				By("putting the v2 layers in the corrent directory structure")
				for _, dir := range createDirs {
					Expect(filepath.Join(layersDir, "buildpack.0", "layer", dir)).To(BeADirectory())
				}

				for _, file := range createFiles {
					Expect(filepath.Join(layersDir, "buildpack.0", "layer", file)).To(BeAnExistingFile())
				}

				By("writing the group metadata in the order the buildpacks should be sourced")
				groupMetadataContents, err := ioutil.ReadFile(groupMetadata)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(groupMetadataContents)).To(Equal(`[[buildpacks]]
  id = "buildpack.0"
  version = ""

[[buildpacks]]
  id = "buildpack.1"
  version = ""
`))
			})
		})

		Context("MoveV2Layers", func() {
			It("moves directories and creates the dst dir if it doesn't exist", func() {
				Expect(supplier.MoveV2Layers(filepath.Join(depsDir, depsIndex), filepath.Join(layersDir, "buildpack.0", "layers.0"))).To(Succeed())
				Expect(filepath.Join(layersDir, "buildpack.0", "layers.0")).To(BeADirectory())
			})
		})

		Context("RenameEnvDir", func() {
			It("renames the env dir to env.build", func() {
				Expect(os.Mkdir(filepath.Join(layersDir, "env"), 0777)).To(Succeed())
				Expect(supplier.RenameEnvDir(layersDir)).To(Succeed())
				Expect(filepath.Join(layersDir, "env.build")).To(BeADirectory())
			})

			It("does nothing when the env dir does NOT exist", func() {
				Expect(supplier.RenameEnvDir(layersDir)).To(Succeed())
				Expect(filepath.Join(layersDir, "env.build")).NotTo(BeADirectory())
			})
		})

		Context("UpdateGroupTOML", func() {
			BeforeEach(func() {
				depsIndex = "1"
				Expect(ioutil.WriteFile(groupMetadata, []byte(`[[buildpacks]]
  id = "org.cloudfoundry.buildpacks.nodejs"
  version = "0.0.2"
[[buildpacks]]
  id = "org.cloudfoundry.buildpacks.npm"
  version = "0.0.3"`), 0777)).To(Succeed())
			})

			It("adds v2 buildpacks to the group.toml", func() {
				Expect(supplier.UpdateGroupTOML("buildpack.0")).To(Succeed())
				groupMetadataContents, err := ioutil.ReadFile(groupMetadata)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(groupMetadataContents)).To(Equal(`[[buildpacks]]
  id = "buildpack.0"
  version = ""

[[buildpacks]]
  id = "org.cloudfoundry.buildpacks.nodejs"
  version = "0.0.2"

[[buildpacks]]
  id = "org.cloudfoundry.buildpacks.npm"
  version = "0.0.3"
`))
			})
		})

		Context("AddFakeCNBBuildpack", func() {
			It("adds the v2 buildpack as a no-op cnb buildpack", func() {
				Expect(os.Setenv("CF_STACK", "cflinuxfs3")).To(Succeed())
				Expect(supplier.AddFakeCNBBuildpack("buildpack.0")).To(Succeed())
				buildpackTOML, err := ioutil.ReadFile(filepath.Join(v3BuildpacksDir, "buildpack.0", "latest", "buildpack.toml"))
				Expect(err).ToNot(HaveOccurred())
				Expect(string(buildpackTOML)).To(Equal(`[buildpack]
  id = "buildpack.0"
  name = "buildpack.0"
  version = "latest"

[[stacks]]
  id = "org.cloudfoundry.stacks.cflinuxfs3"
`))

				Expect(filepath.Join(v3BuildpacksDir, "buildpack.0", "latest", "bin", "build")).To(BeAnExistingFile())
			})
		})
	})

	Describe("Finalizer", func() {
		var (
			finalizer             shims.Finalizer
			profileDir, depsIndex string
		)

		BeforeEach(func() {
			var err error

			depsIndex = "0"

			Expect(os.Setenv("CF_STACK", "some-stack")).To(Succeed())
			profileDir, err = ioutil.TempDir("", "profile")
			Expect(err).NotTo(HaveOccurred())
		})

		JustBeforeEach(func() {
			finalizer = shims.Finalizer{DepsIndex: depsIndex, ProfileDir: profileDir}
		})

		AfterEach(func() {
			Expect(os.Unsetenv("CF_STACK")).To(Succeed())
			Expect(os.RemoveAll(profileDir)).To(Succeed())
		})

		It("writes a profile.d script for the V2 lifecycle to exec which calls the v3-launcher", func() {
			Expect(finalizer.Finalize()).To(Succeed())

			contents, err := ioutil.ReadFile(filepath.Join(profileDir, "0_shim.sh"))
			Expect(err).NotTo(HaveOccurred())

			Expect(string(contents)).To(Equal(`export PACK_STACK_ID="org.cloudfoundry.stacks.some-stack"
export PACK_LAYERS_DIR="$DEPS_DIR"
export PACK_APP_DIR="$HOME"
exec $DEPS_DIR/v3-launcher "$2"
`))
		})
	})

	Describe("Releaser", func() {
		var (
			releaser   shims.Releaser
			v2BuildDir string
			buf        *bytes.Buffer
		)

		BeforeEach(func() {
			var err error

			v2BuildDir, err = ioutil.TempDir("", "build")
			Expect(err).NotTo(HaveOccurred())

			contents := `
			buildpacks = ["some.buildpacks", "some.other.buildpack"]
			[[processes]]
			type = "web"
			command = "npm start"
			`
			Expect(os.MkdirAll(filepath.Join(v2BuildDir, ".cloudfoundry"), 0777)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(v2BuildDir, ".cloudfoundry", "metadata.toml"), []byte(contents), 0666)).To(Succeed())

			buf = &bytes.Buffer{}

			releaser = shims.Releaser{
				MetadataPath: filepath.Join(v2BuildDir, ".cloudfoundry", "metadata.toml"),
				Writer:       buf,
			}
		})

		AfterEach(func() {
			Expect(os.RemoveAll(v2BuildDir)).To(Succeed())
		})

		It("runs with the correct arguments and moves things to the correct place", func() {
			Expect(releaser.Release()).To(Succeed())
			Expect(buf.Bytes()).To(Equal([]byte("default_process_types:\n  web: npm start\n")))
			Expect(filepath.Join(v2BuildDir, ".cloudfoundry", "metadata.toml")).NotTo(BeAnExistingFile())
		})
	})

	Describe("CNBInstaller", func() {
		BeforeEach(func() {
			Expect(os.Setenv("CF_STACK", "cflinuxfs3")).To(Succeed())

			httpmock.Reset()

			contents, err := ioutil.ReadFile("fixtures/bpA.tgz")
			Expect(err).ToNot(HaveOccurred())

			httpmock.RegisterResponder("GET", "https://a-fake-url.com/bpA.tgz",
				httpmock.NewStringResponder(200, string(contents)))

			contents, err = ioutil.ReadFile("fixtures/bpB.tgz")
			Expect(err).ToNot(HaveOccurred())

			httpmock.RegisterResponder("GET", "https://a-fake-url.com/bpB.tgz",
				httpmock.NewStringResponder(200, string(contents)))
		})

		AfterEach(func() {
			Expect(os.Unsetenv("CF_STACK")).To(Succeed())
		})

		It("installs the latest/unique buildpacks from an order.toml", func() {
			tmpDir, err := ioutil.TempDir("", "")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpDir)

			buffer := new(bytes.Buffer)
			logger := libbuildpack.NewLogger(ansicleaner.New(buffer))

			manifest, err := libbuildpack.NewManifest("fixtures", logger, time.Now())
			Expect(err).To(BeNil())

			installer := shims.NewCNBInstaller(manifest)

			Expect(installer.InstallCNBS("fixtures/order.toml", tmpDir)).To(Succeed())
			Expect(filepath.Join(tmpDir, "this.is.a.fake.bpA", "1.0.1", "a.txt")).To(BeAnExistingFile())
			Expect(filepath.Join(tmpDir, "this.is.a.fake.bpB", "1.0.2", "b.txt")).To(BeAnExistingFile())
			Expect(filepath.Join(tmpDir, "this.is.a.fake.bpA", "latest")).To(BeAnExistingFile())
			Expect(filepath.Join(tmpDir, "this.is.a.fake.bpB", "latest")).To(BeAnExistingFile())
		})
	})
})
