package shims_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack/shims"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

//go:generate mockgen -source=detector.go --destination=mocks_detector_shims_test.go --package=shims_test

var _ = Describe("Detector", func() {
	var (
		detector        shims.DefaultDetector
		v3BuildpacksDir string
		v3AppDir        string
		tempDir         string
		v3LifecycleDir  string
		groupMetadata   string
		orderMetadata   string
		planMetadata    string
		mockInstaller   *MockInstaller
		mockCtrl        *gomock.Controller
	)

	BeforeEach(func() {
		var err error
		mockCtrl = gomock.NewController(GinkgoT())
		mockInstaller = NewMockInstaller(mockCtrl)

		tempDir, err = ioutil.TempDir("", "tmp")
		Expect(err).NotTo(HaveOccurred())

		v3AppDir = filepath.Join(tempDir, "cnb-app")

		v3LifecycleDir = filepath.Join(tempDir, "lifecycle-dir")
		Expect(os.MkdirAll(v3LifecycleDir, 0777)).To(Succeed())

		groupMetadata = filepath.Join(tempDir, "metadata", "group.toml")
		orderMetadata = filepath.Join(tempDir, "metadata", "order.toml")
		planMetadata = filepath.Join(tempDir, "metadata", "plan.toml")

		v3BuildpacksDir = filepath.Join(tempDir, "buildpacks")

		detector = shims.DefaultDetector{
			AppDir:          v3AppDir,
			V3BuildpacksDir: v3BuildpacksDir,
			V3LifecycleDir:  v3LifecycleDir,
			OrderMetadata:   orderMetadata,
			GroupMetadata:   groupMetadata,
			PlanMetadata:    planMetadata,
			Installer:       mockInstaller,
		}
	})

	AfterEach(func() {
		Expect(os.RemoveAll(tempDir)).To(Succeed())
	})

	Context("Running Detect", func() {
		It("should Run v3-detector", func() {
			mockInstaller.EXPECT().InstallCNBS(orderMetadata, v3BuildpacksDir)
			mockInstaller.EXPECT().InstallLifecycle(v3LifecycleDir).Do(func(path string) error {
				contents := "#!/usr/bin/env bash\nexit 0\n"
				return ioutil.WriteFile(filepath.Join(path, "detector"), []byte(contents), os.ModePerm)
			})
			Expect(detector.Detect()).ToNot(HaveOccurred())
		})
	})

})
