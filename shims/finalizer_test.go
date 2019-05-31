package shims_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack"

	"github.com/cloudfoundry/libbuildpack/shims"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

//go:generate mockgen -source=finalizer.go --destination=mocks_shims_test.go --package=shims_test
//go:generate mockgen -source=detector.go --destination=mocks_detector_shims_test.go --package=shims_test

var _ = Describe("Finalizer", func() {
	var (
		finalizer    shims.Finalizer
		mockCtrl     *gomock.Controller
		mockDetector *MockDetector
		tempDir,
		v2AppDir,
		v3AppDir,
		v2DepsDir,
		v2CacheDir,
		v3LayersDir,
		v3LauncherDir,
		v3BuildpacksDir,
		orderDir,
		orderMetadata,
		planMetadata,
		groupMetadata,
		profileDir,
		binDir,
		depsIndex string
		finalizeLogger *libbuildpack.Logger
	)

	BeforeEach(func() {
		var err error

		mockCtrl = gomock.NewController(GinkgoT())
		mockDetector = NewMockDetector(mockCtrl)

		depsIndex = "0"

		tempDir, err = ioutil.TempDir("", "tmp")
		Expect(err).NotTo(HaveOccurred())

		v2AppDir = filepath.Join(tempDir, "v2_app")
		Expect(os.MkdirAll(v2AppDir, 0777)).To(Succeed())

		v3AppDir = filepath.Join(tempDir, "v3_app")
		Expect(os.MkdirAll(v3AppDir, 0777)).To(Succeed())

		v2DepsDir = filepath.Join(tempDir, "deps")

		v2CacheDir = filepath.Join(tempDir, "cache")
		Expect(os.MkdirAll(tempDir, 0777)).To(Succeed())

		v3LayersDir = filepath.Join(tempDir, "layers")
		Expect(os.MkdirAll(v3LayersDir, 0777)).To(Succeed())

		v3LauncherDir = filepath.Join(tempDir, "launch")
		Expect(os.MkdirAll(v3LauncherDir, 0777)).To(Succeed())

		v3BuildpacksDir = filepath.Join(tempDir, "cnbs")
		Expect(os.MkdirAll(v3BuildpacksDir, 0777)).To(Succeed())

		orderDir = filepath.Join(tempDir, "order")
		Expect(os.MkdirAll(orderDir, 0777)).To(Succeed())

		orderMetadata = filepath.Join(tempDir, "order.toml")
		planMetadata = filepath.Join(tempDir, "plan.toml")
		groupMetadata = filepath.Join(tempDir, "group.toml")

		profileDir = filepath.Join(tempDir, "profile")
		Expect(os.MkdirAll(profileDir, 0777)).To(Succeed())

		binDir = filepath.Join(tempDir, "bin")
		Expect(os.MkdirAll(binDir, 0777)).To(Succeed())

		Expect(os.Setenv("CF_STACK", "some-stack")).To(Succeed())

		finalizeLogger = &libbuildpack.Logger{}
	})

	JustBeforeEach(func() {
		Expect(os.MkdirAll(filepath.Join(v2DepsDir, depsIndex), 0777)).To(Succeed())

		finalizer = shims.Finalizer{
			V2AppDir:        v2AppDir,
			V3AppDir:        v3AppDir,
			V2DepsDir:       v2DepsDir,
			V2CacheDir:      v2CacheDir,
			V3LayersDir:     v3LayersDir,
			V3LauncherDir:   v3LauncherDir,
			V3BuildpacksDir: v3BuildpacksDir,
			DepsIndex:       depsIndex,
			OrderDir:        orderDir,
			OrderMetadata:   orderMetadata,
			PlanMetadata:    planMetadata,
			GroupMetadata:   groupMetadata,
			ProfileDir:      profileDir,
			V3LifecycleDir:  binDir,
			Detector:        mockDetector,
			Logger:          finalizeLogger,
		}
	})

	AfterEach(func() {
		mockCtrl.Finish()
		Expect(os.Unsetenv("CF_STACK")).To(Succeed())
		Expect(os.RemoveAll(tempDir)).To(Succeed())
	})

	Context("MergeOrderTOMLs", func() {
		var (
			orderFileA, orderFileB, orderFileC string
		)

		BeforeEach(func() {
			orderFileA = filepath.Join(orderDir, "orderA.toml")
			orderFileB = filepath.Join(orderDir, "orderB.toml")
		})

		Context("group A has one group, group B has one groups", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(orderFileA, []byte(generateOrderTOMLGroupString("X", []bp{{id: "bpA", optional: false}, {id: "bpB", optional: true}})), 0777)).To(Succeed())
				Expect(ioutil.WriteFile(orderFileB, []byte(generateOrderTOMLGroupString("Y", []bp{{id: "bpC", optional: false}, {id: "bpD", optional: false}})), 0777)).To(Succeed())
			})

			It("merges the order files", func() {
				Expect(finalizer.MergeOrderTOMLs()).To(Succeed())
				orderTOML, err := ioutil.ReadFile(orderMetadata)
				Expect(err).ToNot(HaveOccurred())

				Expect(string(orderTOML)).To(ContainSubstring(`[[groups]]
  labels = ["X", "Y"]

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpA"
    version = "latest"

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpB"
    version = "latest"
    optional = true

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpC"
    version = "latest"

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpD"
    version = "latest"`))
			})
		})

		Context("group A has one group, group B has two groups", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(orderFileA, []byte(generateOrderTOMLGroupString("X", []bp{{id: "bpA", optional: false}, {id: "bpB", optional: true}})), 0777)).To(Succeed())
				Expect(ioutil.WriteFile(orderFileB, []byte(
					generateOrderTOMLGroupString("Y", []bp{{id: "bpC", optional: false}, {id: "bpD", optional: false}})+"\n"+
						generateOrderTOMLGroupString("Z", []bp{{id: "bpC", optional: false}, {id: "bpE", optional: false}}),
				), 0777)).To(Succeed())
			})

			It("merges the order files", func() {
				Expect(finalizer.MergeOrderTOMLs()).To(Succeed())
				orderTOML, err := ioutil.ReadFile(orderMetadata)
				Expect(err).ToNot(HaveOccurred())

				Expect(string(orderTOML)).To(ContainSubstring(`[[groups]]
  labels = ["X", "Y"]

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpA"
    version = "latest"

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpB"
    version = "latest"
    optional = true

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpC"
    version = "latest"

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpD"
    version = "latest"

[[groups]]
  labels = ["X", "Z"]

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpA"
    version = "latest"

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpB"
    version = "latest"
    optional = true

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpC"
    version = "latest"

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpE"
    version = "latest"`))
			})
		})

		Context("group A has two group, group B has two groups", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(orderFileA, []byte(
					generateOrderTOMLGroupString("X1", []bp{{id: "bpA", optional: false}, {id: "bpB", optional: false}})+
						"\n"+
						generateOrderTOMLGroupString("X2", []bp{{id: "bpA", optional: false}, {id: "bpC", optional: false}}),
				), 0777)).To(Succeed())
				Expect(ioutil.WriteFile(orderFileB, []byte(
					generateOrderTOMLGroupString("Y1", []bp{{id: "bpD", optional: false}, {id: "bpE", optional: false}})+
						"\n"+
						generateOrderTOMLGroupString("Y2", []bp{{id: "bpD", optional: false}, {id: "bpF", optional: false}}),
				), 0777)).To(Succeed())
			})

			It("merges the order files", func() {
				Expect(finalizer.MergeOrderTOMLs()).To(Succeed())
				orderTOML, err := ioutil.ReadFile(orderMetadata)
				Expect(err).ToNot(HaveOccurred())

				Expect(string(orderTOML)).To(ContainSubstring(`[[groups]]
  labels = ["X1", "Y1"]

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpA"
    version = "latest"

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpB"
    version = "latest"

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpD"
    version = "latest"

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpE"
    version = "latest"

[[groups]]
  labels = ["X1", "Y2"]

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpA"
    version = "latest"

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpB"
    version = "latest"

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpD"
    version = "latest"

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpF"
    version = "latest"

[[groups]]
  labels = ["X2", "Y1"]

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpA"
    version = "latest"

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpC"
    version = "latest"

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpD"
    version = "latest"

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpE"
    version = "latest"

[[groups]]
  labels = ["X2", "Y2"]

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpA"
    version = "latest"

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpC"
    version = "latest"

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpD"
    version = "latest"

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpF"
    version = "latest"`))
			})
		})

		Context("group A has one group, group B has one groups, groupC has two groups", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(orderFileA, []byte(generateOrderTOMLGroupString("X", []bp{{id: "bpA", optional: false}, {id: "bpB", optional: true}})), 0777)).To(Succeed())
				Expect(ioutil.WriteFile(orderFileB, []byte(generateOrderTOMLGroupString("Y", []bp{{id: "bpC", optional: false}, {id: "bpD", optional: false}})), 0777)).To(Succeed())
				orderFileC = filepath.Join(orderDir, "orderC.toml")
				Expect(ioutil.WriteFile(orderFileC, []byte(
					generateOrderTOMLGroupString("Z1", []bp{{id: "bpE", optional: false}, {id: "bpF", optional: false}})+
						"\n"+
						generateOrderTOMLGroupString("Z2", []bp{{id: "bpE", optional: false}, {id: "bpG", optional: false}}),
				), 0777)).To(Succeed())
			})

			It("merges the order files", func() {
				Expect(finalizer.MergeOrderTOMLs()).To(Succeed())
				orderTOML, err := ioutil.ReadFile(orderMetadata)
				Expect(err).ToNot(HaveOccurred())

				Expect(string(orderTOML)).To(ContainSubstring(`[[groups]]
  labels = ["X", "Y", "Z1"]

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpA"
    version = "latest"

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpB"
    version = "latest"
    optional = true

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpC"
    version = "latest"

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpD"
    version = "latest"

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpE"
    version = "latest"

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpF"
    version = "latest"

[[groups]]
  labels = ["X", "Y", "Z2"]

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpA"
    version = "latest"

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpB"
    version = "latest"
    optional = true

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpC"
    version = "latest"

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpD"
    version = "latest"

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpE"
    version = "latest"

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpG"
    version = "latest"
`))
			})
		})

		Context("group A has one group, group B has one group with a duplicate buildpack", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(orderFileA, []byte(generateOrderTOMLGroupString("X", []bp{{id: "bpA", optional: false}, {id: "bpB", optional: false}})), 0777)).To(Succeed())
				Expect(ioutil.WriteFile(orderFileB, []byte(generateOrderTOMLGroupString("Y", []bp{{id: "bpA", optional: false}, {id: "bpC", optional: false}})), 0777)).To(Succeed())
			})

			It("merges the order files", func() {
				Expect(finalizer.MergeOrderTOMLs()).To(Succeed())
				orderTOML, err := ioutil.ReadFile(orderMetadata)
				Expect(err).ToNot(HaveOccurred())

				Expect(string(orderTOML)).To(ContainSubstring(`[[groups]]
  labels = ["X", "Y"]

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpA"
    version = "latest"

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpB"
    version = "latest"

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpC"
    version = "latest"`))
			})
		})
	})

	Context("RunV3Detect", func() {
		It("runs detection when group or plan metadata does not exist", func() {
			mockDetector.
				EXPECT().
				RunLifecycleDetect()
			Expect(finalizer.RunV3Detect()).To(Succeed())
		})

		It("does NOT run detection when group and plan metadata exists", func() {
			Expect(ioutil.WriteFile(groupMetadata, []byte(""), 0666)).To(Succeed())
			Expect(ioutil.WriteFile(planMetadata, []byte(""), 0666)).To(Succeed())

			mockDetector.
				EXPECT().
				RunLifecycleDetect().
				Times(0)
			Expect(finalizer.RunV3Detect()).To(Succeed())
		})
	})

	Context("IncludePreviousV2Buildpacks", func() {
		var (
			createDirs, createFiles []string
		)

		BeforeEach(func() {
			depsIndex = "2"
			createDirs = []string{"bin", "lib"}
			createFiles = []string{"config.yml"}
			for _, dir := range createDirs {
				Expect(os.MkdirAll(filepath.Join(v2DepsDir, "0", dir), 0777)).To(Succeed())
			}

			for _, file := range createFiles {
				Expect(ioutil.WriteFile(filepath.Join(v2DepsDir, "0", file), []byte(file), 0666)).To(Succeed())
			}

			Expect(ioutil.WriteFile(groupMetadata, []byte(`[[buildpacks]]
  id = "buildpack.1"
  version = ""
[[buildpacks]]
  id = "buildpack.2"
  version = ""`), 0666)).To(Succeed())
			Expect(ioutil.WriteFile(planMetadata, []byte(""), 0666)).To(Succeed())
		})

		It("copies v2 layers and metadata where v3 lifecycle expects them for build and launch", func() {
			By("not failing if a layer has already been moved")
			Expect(finalizer.IncludePreviousV2Buildpacks()).To(Succeed())

			By("putting the v2 layers in the corrent directory structure")
			for _, dir := range createDirs {
				Expect(filepath.Join(v3LayersDir, "buildpack.0", "layer", dir)).To(BeADirectory())
			}

			for _, file := range createFiles {
				Expect(filepath.Join(v3LayersDir, "buildpack.0", "layer", file)).To(BeAnExistingFile())
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

[[buildpacks]]
  id = "buildpack.2"
  version = ""
`))
		})
	})

	Context("RestoreV3Cache", func() {
		BeforeEach(func() {
			cloudfoundryV3Cache := filepath.Join(v2CacheDir, "cnb")
			testLayers := filepath.Join(cloudfoundryV3Cache, "org.cloudfoundry.generic.buildpack")
			Expect(os.MkdirAll(cloudfoundryV3Cache, 0777)).To(Succeed())
			Expect(os.MkdirAll(filepath.Join(testLayers, "layer"), 0777)).To(Succeed())
			Expect(os.MkdirAll(filepath.Join(testLayers, "anotherLayer"), 0777)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(testLayers, "anotherLayer", "cachedContents"), []byte("cached contents"), 0666)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(testLayers, "anotherLayer", "anotherLayer.toml"), []byte("cache=true"), 0666)).To(Succeed())
		})

		It("should restore cache before building", func() {
			restoredLayers := filepath.Join(finalizer.V3LayersDir, "org.cloudfoundry.generic.buildpack")
			Expect(finalizer.RestoreV3Cache()).ToNot(HaveOccurred())
			Expect(filepath.Join(restoredLayers, "layer")).To(BeADirectory())
			Expect(filepath.Join(restoredLayers, "anotherLayer")).To(BeADirectory())
			Expect(filepath.Join(restoredLayers, "anotherLayer", "cachedContents")).To(BeAnExistingFile())
			contents, err := ioutil.ReadFile(filepath.Join(restoredLayers, "anotherLayer", "cachedContents"))
			Expect(err).ToNot(HaveOccurred())
			Expect(contents).To(ContainSubstring("cached contents"))
		})
	})

	Context("MoveV3Layers", func() {
		BeforeEach(func() {
			Expect(os.MkdirAll(filepath.Join(v3LayersDir, "config"), 0777)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(v3LayersDir, "config", "metadata.toml"), []byte(""), 0666)).To(Succeed())

			Expect(os.MkdirAll(filepath.Join(v3LayersDir, "layers"), 0777)).To(Succeed())
			Expect(os.MkdirAll(filepath.Join(v3LayersDir, "anotherLayers", "innerLayer"), 0777)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(v3LayersDir, "anotherLayers", "innerLayer.toml"), []byte("cache=true"), 0666)).To(Succeed())
		})

		It("moves the layers to deps dir and metadata to build dir", func() {
			Expect(finalizer.MoveV3Layers()).To(Succeed())
			Expect(filepath.Join(v2AppDir, ".cloudfoundry", "metadata.toml")).To(BeAnExistingFile())
			Expect(filepath.Join(v2DepsDir, "layers")).To(BeAnExistingFile())
			Expect(filepath.Join(v2DepsDir, "anotherLayers")).To(BeAnExistingFile())
		})

		It("copies cacheable layers to the cache/cnb directory", func() {
			Expect(filepath.Join(v2CacheDir, "cnb")).ToNot(BeADirectory())
			Expect(finalizer.MoveV3Layers()).To(Succeed())
			Expect(filepath.Join(v2CacheDir, "cnb", "anotherLayers", "innerLayer")).To(BeADirectory())
		})
	})

	Context("MoveV2Layers", func() {
		It("moves directories and creates the dst dir if it doesn't exist", func() {
			Expect(finalizer.MoveV2Layers(filepath.Join(v2DepsDir, depsIndex), filepath.Join(v3LayersDir, "buildpack.0", "layers.0"))).To(Succeed())
			Expect(filepath.Join(v3LayersDir, "buildpack.0", "layers.0")).To(BeADirectory())
		})
	})

	Context("RenameEnvDir", func() {
		It("renames the env dir to env.build", func() {
			Expect(os.Mkdir(filepath.Join(v3LayersDir, "env"), 0777)).To(Succeed())
			Expect(finalizer.RenameEnvDir(v3LayersDir)).To(Succeed())
			Expect(filepath.Join(v3LayersDir, "env.build")).To(BeADirectory())
		})

		It("does nothing when the env dir does NOT exist", func() {
			Expect(finalizer.RenameEnvDir(v3LayersDir)).To(Succeed())
			Expect(filepath.Join(v3LayersDir, "env.build")).NotTo(BeADirectory())
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
			Expect(finalizer.UpdateGroupTOML("buildpack.0")).To(Succeed())
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

	Context("In V3 Layers Dir", func() {
		var (
			testLayers            string
			Dep1LayerMetadataPath string
			Dep2LayerMetadataPath string
		)

		JustBeforeEach(func() {
			testLayers = filepath.Join(finalizer.V3LayersDir, "org.cloudfoundry.generic.buildpack")
			Expect(os.MkdirAll(testLayers, os.ModePerm)).To(Succeed())
			Dep1LayerMetadataPath = filepath.Join(testLayers, "dep1.toml")
			Dep2LayerMetadataPath = filepath.Join(testLayers, "dep2.toml")
			Dep1LayerPath := filepath.Join(testLayers, "dep1")
			Dep2LayerPath := filepath.Join(testLayers, "dep2")
			Expect(os.MkdirAll(Dep2LayerPath, os.ModePerm)).To(Succeed())
			Expect(os.MkdirAll(Dep1LayerPath, os.ModePerm)).To(Succeed())

			Expect(ioutil.WriteFile(Dep1LayerMetadataPath, []byte(`launch = true
			build = false
			cache = true

			[metadata]
			extradata = "shamoo"`), 0777)).To(Succeed())

			Expect(ioutil.WriteFile(Dep2LayerMetadataPath, []byte(`launch = true
			build = true
			cache = true
			[metadata]
			extradata = "shamwow"`), 0777)).To(Succeed())
		})

		It("can read layer.toml", func() {
			dep1Metadata, err := finalizer.ReadLayerMetadata(Dep1LayerMetadataPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(dep1Metadata.Launch).To(Equal(true))
			Expect(dep1Metadata.Build).To(Equal(false))
			Expect(dep1Metadata.Cache).To(Equal(true))

			dep2Metadata, err := finalizer.ReadLayerMetadata(Dep2LayerMetadataPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(dep2Metadata.Launch).To(Equal(true))
			Expect(dep2Metadata.Build).To(Equal(true))
			Expect(dep2Metadata.Cache).To(Equal(true))
		})

		It("can move layer to cache if needed", func() {
			Expect(finalizer.MoveV3Layers()).To(Succeed())
			layersCacheDir := filepath.Join(v2CacheDir, "cnb", "org.cloudfoundry.generic.buildpack")
			Expect(filepath.Join(layersCacheDir, "dep1")).To(BeADirectory())
			Expect(filepath.Join(layersCacheDir, "dep2")).To(BeADirectory())

		})
	})

	Context("AddFakeCNBBuildpack", func() {
		It("adds the v2 buildpack as a no-op cnb buildpack", func() {
			Expect(os.Setenv("CF_STACK", "cflinuxfs3")).To(Succeed())
			Expect(finalizer.AddFakeCNBBuildpack("buildpack.0")).To(Succeed())
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

	Context("WriteProfileLaunch", func() {
		It("writes a profile script that execs the v3 launcher", func() {
			Expect(finalizer.WriteProfileLaunch()).To(Succeed())
			contents, err := ioutil.ReadFile(filepath.Join(profileDir, shims.V3LaunchScript))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(Equal(fmt.Sprintf(`export CNB_STACK_ID="org.cloudfoundry.stacks.%s"
export CNB_LAYERS_DIR="$DEPS_DIR"
export CNB_APP_DIR="$HOME"
exec $HOME/.cloudfoundry/%s "$2"
`, os.Getenv("CF_STACK"), shims.V3Launcher)))
		})
	})
})

type bp struct {
	id       string
	optional bool
}

func generateOrderTOMLGroupString(label string, bps []bp) string {
	orderToml := fmt.Sprintf(`[[groups]]
	 labels = ["%s"]`, label)

	for _, bp := range bps {
		orderToml +=
			fmt.Sprintf(`

  [[groups.buildpacks]]
    id = "this.is.a.fake.%s"
    version = "latest"`, bp.id)
		if bp.optional {
			orderToml += `
    optional = true`
		}
	}

	return orderToml
}
