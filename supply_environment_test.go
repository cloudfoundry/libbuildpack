package libbuildpack_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	bp "github.com/cloudfoundry/libbuildpack"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Supply Environment", func() {
	var (
		buildDir string
		depsDir  string
		err      error
	)

	BeforeEach(func() {
		buildDir, err = ioutil.TempDir("", "build")
		Expect(err).To(BeNil())

		depsDir, err = ioutil.TempDir("", "deps")
		Expect(err).To(BeNil())

		err = os.MkdirAll(filepath.Join(depsDir, "00", "bin"), 0755)
		Expect(err).To(BeNil())

		err = os.MkdirAll(filepath.Join(depsDir, "01", "bin"), 0755)
		Expect(err).To(BeNil())

		err = os.MkdirAll(filepath.Join(depsDir, "01", "ld_library_path"), 0755)
		Expect(err).To(BeNil())

		err = os.MkdirAll(filepath.Join(depsDir, "02", "ld_library_path"), 0755)
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		err = os.RemoveAll(buildDir)
		Expect(err).To(BeNil())

		err = os.RemoveAll(depsDir)
		Expect(err).To(BeNil())
	})

	Describe("SetEnvironmentFromSupply", func() {
		var (
			oldPath          string
			oldLdLibraryPath string
		)

		BeforeEach(func() {
			oldPath = os.Getenv("PATH")
			oldLdLibraryPath = os.Getenv("PATH")
		})

		AfterEach(func() {
			err = os.Setenv("PATH", oldPath)
			Expect(err).To(BeNil())

			err = os.Setenv("LD_LIBRARY_PATH", oldPath)
			Expect(err).To(BeNil())
		})

		It("sets PATH based on the supplied deps", func() {
			err = bp.SetEnvironmentFromSupply(depsDir)
			Expect(err).To(BeNil())

			newPath := os.Getenv("PATH")
			Expect(newPath).To(Equal(fmt.Sprintf("%s/01/bin:%s/00/bin:%s", depsDir, depsDir, oldPath)))
		})

		It("sets LD_LIBRARY_PATH based on the supplied deps", func() {
			err = bp.SetEnvironmentFromSupply(depsDir)
			Expect(err).To(BeNil())

			newPath := os.Getenv("LD_LIBRARY_PATH")
			Expect(newPath).To(Equal(fmt.Sprintf("%s/02/ld_library_path:%s/01/ld_library_path:%s", depsDir, depsDir, oldLdLibraryPath)))
		})
	})

	Describe("WriteProfileDFromSupply", func() {
		It("writes a .profile.d script allowing the runtime container to use the supplied deps", func() {
			err = bp.WriteProfileDFromSupply(depsDir, buildDir)
			Expect(err).To(BeNil())

			contents, err := ioutil.ReadFile(filepath.Join(buildDir, ".profile.d", "00-multi-supply.sh"))
			Expect(err).To(BeNil())

			Expect(string(contents)).To(ContainSubstring("export PATH=$DEPS_DIR/01/bin:$DEPS_DIR/00/bin:$PATH"))
			Expect(string(contents)).To(ContainSubstring("export LD_LIBRARY_PATH=$DEPS_DIR/02/ld_library_path:$DEPS_DIR/01/ld_library_path:$LD_LIBRARY_PATH"))
		})
	})
})
