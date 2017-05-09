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

		err = os.MkdirAll(filepath.Join(depsDir, "01", "lib"), 0755)
		Expect(err).To(BeNil())

		err = os.MkdirAll(filepath.Join(depsDir, "02", "lib"), 0755)
		Expect(err).To(BeNil())

		err = os.MkdirAll(filepath.Join(depsDir, "03", "include"), 0755)
		Expect(err).To(BeNil())

		err = os.MkdirAll(filepath.Join(depsDir, "04", "pkgconfig"), 0755)
		Expect(err).To(BeNil())

		err = os.MkdirAll(filepath.Join(depsDir, "05", "env"), 0755)
		Expect(err).To(BeNil())

		err = ioutil.WriteFile(filepath.Join(depsDir, "05", "env", "ENV_VAR"), []byte("value"), 0644)
		Expect(err).To(BeNil())

		err = os.MkdirAll(filepath.Join(depsDir, "00", "profile.d"), 0755)
		Expect(err).To(BeNil())

		err = ioutil.WriteFile(filepath.Join(depsDir, "00", "profile.d", "supplied-script.sh"), []byte("first"), 0644)
		Expect(err).To(BeNil())

		err = os.MkdirAll(filepath.Join(depsDir, "01", "profile.d"), 0755)
		Expect(err).To(BeNil())

		err = ioutil.WriteFile(filepath.Join(depsDir, "01", "profile.d", "supplied-script.sh"), []byte("second"), 0644)
		Expect(err).To(BeNil())

		err = ioutil.WriteFile(filepath.Join(depsDir, "some-file.yml"), []byte("things"), 0644)
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		err = os.RemoveAll(buildDir)
		Expect(err).To(BeNil())

		err = os.RemoveAll(depsDir)
		Expect(err).To(BeNil())
	})

	Describe("SetStagingEnvironment", func() {
		var envVars = map[string]string{}

		BeforeEach(func() {
			vars := []string{"PATH", "LD_LIBRARY_PATH", "INCLUDE_PATH", "CPATH", "CPPPATH", "PKG_CONFIG_PATH", "ENV_VAR"}

			for _, envVar := range vars {
				envVars[envVar] = os.Getenv(envVar)
				os.Setenv(envVar, "existing_"+envVar)
			}
		})

		AfterEach(func() {
			for key, val := range envVars {
				err = os.Setenv(key, val)
				Expect(err).To(BeNil())
			}
		})

		It("sets PATH based on the supplied deps", func() {
			err = bp.SetStagingEnvironment(depsDir)
			Expect(err).To(BeNil())

			newPath := os.Getenv("PATH")
			Expect(newPath).To(Equal(fmt.Sprintf("%s/01/bin:%s/00/bin:existing_PATH", depsDir, depsDir)))
		})

		It("sets LD_LIBRARY_PATH based on the supplied deps", func() {
			err = bp.SetStagingEnvironment(depsDir)
			Expect(err).To(BeNil())

			newPath := os.Getenv("LD_LIBRARY_PATH")
			Expect(newPath).To(Equal(fmt.Sprintf("%s/02/lib:%s/01/lib:existing_LD_LIBRARY_PATH", depsDir, depsDir)))
		})

		It("sets INCLUDE_PATH based on the supplied deps", func() {
			err = bp.SetStagingEnvironment(depsDir)
			Expect(err).To(BeNil())

			newPath := os.Getenv("INCLUDE_PATH")
			Expect(newPath).To(Equal(fmt.Sprintf("%s/03/include:existing_INCLUDE_PATH", depsDir)))
		})

		It("sets CPATH based on the supplied deps", func() {
			err = bp.SetStagingEnvironment(depsDir)
			Expect(err).To(BeNil())

			newPath := os.Getenv("CPATH")
			Expect(newPath).To(Equal(fmt.Sprintf("%s/03/include:existing_CPATH", depsDir)))
		})

		It("sets CPPPATH based on the supplied deps", func() {
			err = bp.SetStagingEnvironment(depsDir)
			Expect(err).To(BeNil())

			newPath := os.Getenv("CPPPATH")
			Expect(newPath).To(Equal(fmt.Sprintf("%s/03/include:existing_CPPPATH", depsDir)))
		})

		It("sets PKG_CONFIG_PATH based on the supplied deps", func() {
			err = bp.SetStagingEnvironment(depsDir)
			Expect(err).To(BeNil())

			newPath := os.Getenv("PKG_CONFIG_PATH")
			Expect(newPath).To(Equal(fmt.Sprintf("%s/04/pkgconfig:existing_PKG_CONFIG_PATH", depsDir)))
		})

		It("sets environment variables from the env/ dir", func() {
			err = bp.SetStagingEnvironment(depsDir)
			Expect(err).To(BeNil())

			newPath := os.Getenv("ENV_VAR")
			Expect(newPath).To(Equal("value"))
		})

		Context("relevant env variable is empty", func() {
			BeforeEach(func() {
				for key, _ := range envVars {
					os.Setenv(key, "")
				}
			})
			It("sets PATH based on the supplied deps", func() {
				err = bp.SetStagingEnvironment(depsDir)
				Expect(err).To(BeNil())

				newPath := os.Getenv("PATH")
				Expect(newPath).To(Equal(fmt.Sprintf("%s/01/bin:%s/00/bin", depsDir, depsDir)))
			})

			It("sets LD_LIBRARY_PATH based on the supplied deps", func() {
				err = bp.SetStagingEnvironment(depsDir)
				Expect(err).To(BeNil())

				newPath := os.Getenv("LD_LIBRARY_PATH")
				Expect(newPath).To(Equal(fmt.Sprintf("%s/02/lib:%s/01/lib", depsDir, depsDir)))
			})

			It("sets INCLUDE_PATH based on the supplied deps", func() {
				err = bp.SetStagingEnvironment(depsDir)
				Expect(err).To(BeNil())

				newPath := os.Getenv("INCLUDE_PATH")
				Expect(newPath).To(Equal(fmt.Sprintf("%s/03/include", depsDir)))
			})

			It("sets CPATH based on the supplied deps", func() {
				err = bp.SetStagingEnvironment(depsDir)
				Expect(err).To(BeNil())

				newPath := os.Getenv("CPATH")
				Expect(newPath).To(Equal(fmt.Sprintf("%s/03/include", depsDir)))
			})

			It("sets CPPPATH based on the supplied deps", func() {
				err = bp.SetStagingEnvironment(depsDir)
				Expect(err).To(BeNil())

				newPath := os.Getenv("CPPPATH")
				Expect(newPath).To(Equal(fmt.Sprintf("%s/03/include", depsDir)))
			})

			It("sets PKG_CONFIG_PATH based on the supplied deps", func() {
				err = bp.SetStagingEnvironment(depsDir)
				Expect(err).To(BeNil())

				newPath := os.Getenv("PKG_CONFIG_PATH")
				Expect(newPath).To(Equal(fmt.Sprintf("%s/04/pkgconfig", depsDir)))
			})
		})
	})

	Describe("SetLaunchEnvironment", func() {
		It("writes a .profile.d script allowing the runtime container to use the supplied deps", func() {
			err = bp.SetLaunchEnvironment(depsDir, buildDir)
			Expect(err).To(BeNil())

			contents, err := ioutil.ReadFile(filepath.Join(buildDir, ".profile.d", "000_multi-supply.sh"))
			Expect(err).To(BeNil())

			Expect(string(contents)).To(ContainSubstring(`export PATH=$DEPS_DIR/01/bin:$DEPS_DIR/00/bin$([[ ! -z "${PATH:-}" ]] && echo ":$PATH")`))
			Expect(string(contents)).To(ContainSubstring(`export LD_LIBRARY_PATH=$DEPS_DIR/02/lib:$DEPS_DIR/01/lib$([[ ! -z "${LD_LIBRARY_PATH:-}" ]] && echo ":$LD_LIBRARY_PATH")`))
		})

		It("copies scripts from <deps-dir>/<idx>/profile.d to the .profile.d directory, prepending <idx>", func() {
			err = bp.SetLaunchEnvironment(depsDir, buildDir)
			Expect(err).To(BeNil())

			contents, err := ioutil.ReadFile(filepath.Join(buildDir, ".profile.d", "00_supplied-script.sh"))
			Expect(err).To(BeNil())

			Expect(string(contents)).To(Equal("first"))

			contents, err = ioutil.ReadFile(filepath.Join(buildDir, ".profile.d", "01_supplied-script.sh"))
			Expect(err).To(BeNil())

			Expect(string(contents)).To(Equal("second"))
		})
	})
})
