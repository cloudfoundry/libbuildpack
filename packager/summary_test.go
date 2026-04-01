package packager_test

import (
	"strings"

	"github.com/cloudfoundry/libbuildpack/packager"
	httpmock "github.com/jarcoal/httpmock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Packager", func() {
	var (
		buildpackDir string
	)
	BeforeEach(func() {
		buildpackDir = "./fixtures/good"

		httpmock.Reset()
	})

	Describe("Summary", func() {
		It("Renders tables of dependencies", func() {
			s, e := packager.Summary(buildpackDir)
			Expect(e).NotTo(HaveOccurred())
			Expect(s, e).To(Equal(`
Packaged binaries:

| name | version | cf_stacks |
|-|-|-|
| ruby | 1.2.3 | cflinuxfs2 |
| ruby | 1.2.3 | cflinuxfs3 |

Default binary versions:

| name | version |
|-|-|
| ruby | 1.2.3 |
`))
		})

		Context("modules exist", func() {
			BeforeEach(func() {
				buildpackDir = "./fixtures/sub_dependencies"
			})
			It("Renders tables of dependencies (including modules)", func() {
				Expect(packager.Summary(buildpackDir)).To(Equal(`
Packaged binaries:

| name | version | cf_stacks | modules |
|-|-|-|-|
| nginx | 1.7.3 | cflinuxfs2 |  |
| php | 1.6.1 | cflinuxfs2 | gearman, geoip, zlib |
`))
			})
		})

		Context("no dependencies", func() {
			BeforeEach(func() {
				buildpackDir = "./fixtures/no_dependencies"
			})
			It("Produces no output", func() {
				Expect(packager.Summary(buildpackDir)).To(Equal(""))
			})
		})

		Context("manifest has packaging_profiles", func() {
			BeforeEach(func() {
				buildpackDir = "./fixtures/with_profiles"
			})
			It("prints a Packaging profiles section with sorted profile names and descriptions", func() {
				s, err := packager.Summary(buildpackDir)
				Expect(err).NotTo(HaveOccurred())
				Expect(s).To(ContainSubstring("Packaging profiles:"))
				Expect(s).To(ContainSubstring("minimal"))
				Expect(s).To(ContainSubstring("Core deps only. No agents or profilers."))
				Expect(s).To(ContainSubstring("no-profiler"))
				Expect(s).To(ContainSubstring("Core and agents only. No profilers."))
				// minimal sorts before no-profiler
				Expect(strings.Index(s, "minimal")).To(BeNumerically("<", strings.Index(s, "no-profiler")))
			})
		})
	})
})
