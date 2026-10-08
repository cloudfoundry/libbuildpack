package libbuildpack_test

import (
	"fmt"

	bp "github.com/cloudfoundry/libbuildpack"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("versions", func() {

	Describe("FindMatchingVersion with 4-part versions", func() {
		It("matches a minor-line constraint against a 4-part version", func() {
			ver, err := bp.FindMatchingVersion("21.x", []string{"21.0.12.1"})
			Expect(err).NotTo(HaveOccurred())
			Expect(ver).To(Equal("21.0.12.1"))
		})

		It("returns the original 4-part string, not the normalized form", func() {
			ver, err := bp.FindMatchingVersion("21.x", []string{"21.0.11.2", "21.0.12.1"})
			Expect(err).NotTo(HaveOccurred())
			Expect(ver).To(Equal("21.0.12.1"))
		})

		It("matches 3-part constraint against mixed 3-part and 4-part list", func() {
			ver, err := bp.FindMatchingVersion("21.x", []string{"17.0.20.1", "21.0.12.1", "25.0.4.1"})
			Expect(err).NotTo(HaveOccurred())
			Expect(ver).To(Equal("21.0.12.1"))
		})

		It("resolves the highest 4-part version matching a minor-line constraint", func() {
			ver, err := bp.FindMatchingVersion("21.0.x", []string{"21.0.11.2", "21.0.12.1", "21.0.12.3"})
			Expect(err).NotTo(HaveOccurred())
			Expect(ver).To(Equal("21.0.12.3"))
		})

		It("prefers 4-part over 3-part when both share the same prefix", func() {
			ver, err := bp.FindMatchingVersion("21.x", []string{"21.0.12.1", "21.0.12"})
			Expect(err).NotTo(HaveOccurred())
			Expect(ver).To(Equal("21.0.12.1"))
		})

		It("3-part constraint matches 4-part version with same prefix", func() {
			ver, err := bp.FindMatchingVersion("21.0.12", []string{"21.0.12.1"})
			Expect(err).NotTo(HaveOccurred())
			Expect(ver).To(Equal("21.0.12.1"))
		})

		It("matches an exact 4-part version constraint via string equality", func() {
			ver, err := bp.FindMatchingVersion("21.0.12.1", []string{"21.0.11.2", "21.0.12.1"})
			Expect(err).NotTo(HaveOccurred())
			Expect(ver).To(Equal("21.0.12.1"))
		})

		It("returns error for exact 4-part constraint when version is absent", func() {
			_, err := bp.FindMatchingVersion("21.0.12.1", []string{"21.0.12.2"})
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("FindMatchingVersion regression: pre-release versions with dot-separated identifiers", func() {
		It("does not truncate pre-release version strings", func() {
			ver, err := bp.FindMatchingVersion("~8.0.x-0", []string{
				"8.0.100-preview.1.23115.2",
				"8.0.100-preview.7.23376.3",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(ver).To(Equal("8.0.100-preview.7.23376.3"))
		})

		It("picks the latest dotnet preview regardless of input order", func() {
			for _, versions := range [][]string{
				{"8.0.100-preview.7.23376.3", "8.0.100-preview.1.23115.2"},
				{"8.0.100-preview.1.23115.2", "8.0.100-preview.7.23376.3"},
			} {
				ver, err := bp.FindMatchingVersion("~8.0.x-0", versions)
				Expect(err).NotTo(HaveOccurred())
				Expect(ver).To(Equal("8.0.100-preview.7.23376.3"))
			}
		})

		It("handles mixed pre-release and stable versions correctly", func() {
			ver, err := bp.FindMatchingVersion("8.0.x", []string{
				"8.0.1",
				"8.0.100-preview.7.23376.3",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(ver).To(Equal("8.0.100-preview.7.23376.3"))
		})
	})

	Describe("FindMatchingVersions regression", func() {
		It("handles build metadata (Liberica-style openjdk versions)", func() {
			vers, err := bp.FindMatchingVersions("21.x", []string{"21.0.11+9", "21.0.12+10", "17.0.20+10"})
			Expect(err).NotTo(HaveOccurred())
			Expect(vers).To(Equal([]string{"21.0.11+9", "21.0.12+10"}))
		})

		It("matches an exact version with build metadata", func() {
			vers, err := bp.FindMatchingVersions("21.0.12+10", []string{"21.0.11+9", "21.0.12+10"})
			Expect(err).NotTo(HaveOccurred())
			Expect(vers).To(Equal([]string{"21.0.12+10"}))
		})

		It("matches an exact 3-part version only", func() {
			vers, err := bp.FindMatchingVersions("1.2.3", []string{"1.2.3", "1.2.4", "1.3.0"})
			Expect(err).NotTo(HaveOccurred())
			Expect(vers).To(Equal([]string{"1.2.3"}))
		})

		It("supports tilde constraints", func() {
			vers, err := bp.FindMatchingVersions("~1.2.0", []string{"1.2.3", "1.2.4", "1.3.0"})
			Expect(err).NotTo(HaveOccurred())
			Expect(vers).To(Equal([]string{"1.2.3", "1.2.4"}))
		})

		It("supports caret constraints", func() {
			vers, err := bp.FindMatchingVersions("^1.2.0", []string{"1.2.3", "1.2.4", "2.0.0"})
			Expect(err).NotTo(HaveOccurred())
			Expect(vers).To(Equal([]string{"1.2.3", "1.2.4"}))
		})

		It("supports or-constraints", func() {
			vers, err := bp.FindMatchingVersions("1.2.x || 2.x", []string{"1.2.3", "1.3.0", "2.0.0"})
			Expect(err).NotTo(HaveOccurred())
			Expect(vers).To(Equal([]string{"1.2.3", "2.0.0"}))
		})
	})

	Describe("FindMatchingVersion", func() {
		var versions []string

		BeforeEach(func() {
			versions = []string{"1.2.3", "1.2.4", "1.2.2", "1.3.3", "1.3.4", "1.3.2", "2.0.0"}
		})

		It("returns the greatest version", func() {
			ver, err := bp.FindMatchingVersion("x", versions)
			Expect(err).NotTo(HaveOccurred())
			Expect(ver).To(Equal("2.0.0"))
		})

		It("returns the greatest version in a minor line", func() {
			ver, err := bp.FindMatchingVersion("1.x", versions)
			Expect(err).NotTo(HaveOccurred())
			Expect(ver).To(Equal("1.3.4"))
		})

		It("returns the greatest version in a patch line", func() {
			ver, err := bp.FindMatchingVersion("1.2.x", versions)
			Expect(err).NotTo(HaveOccurred())
			Expect(ver).To(Equal("1.2.4"))
		})

		It("returns the greatest version less than the above", func() {
			ver, err := bp.FindMatchingVersion(">=1.2.0, <1.2.4", versions)
			Expect(err).NotTo(HaveOccurred())
			Expect(ver).To(Equal("1.2.3"))
		})

		It("returns the greatest version less than the above (without comma)", func() {
			ver, err := bp.FindMatchingVersion(">=1.2.0 <1.2.4", versions)
			Expect(err).NotTo(HaveOccurred())
			Expect(ver).To(Equal("1.2.3"))
		})

		It("returns the greatest version less or equal than the above (without comma)", func() {
			ver, err := bp.FindMatchingVersion(">1.2.0 <=1.2.4", versions)
			Expect(err).NotTo(HaveOccurred())
			Expect(ver).To(Equal("1.2.4"))
		})

		It("returns an error if nothing matches", func() {
			_, err := bp.FindMatchingVersion("1.4.x", versions)
			Expect(err).To(MatchError(fmt.Sprintf("no match found for 1.4.x in %v", versions)))
		})
	})
	Describe("FindMatchingVersions", func() {
		var versions []string

		BeforeEach(func() {
			versions = []string{"1.2.3", "1.2.4", "1.2.2", "1.3.3", "1.3.4", "1.3.2", "2.0.0"}
		})

		It("returns every version for x, sorted", func() {
			vers, err := bp.FindMatchingVersions("x", versions)
			Expect(err).NotTo(HaveOccurred())
			Expect(vers).To(Equal([]string{"1.2.2", "1.2.3", "1.2.4", "1.3.2", "1.3.3", "1.3.4", "2.0.0"}))
		})

		It("returns all versions in a minor line, sorted", func() {
			vers, err := bp.FindMatchingVersions("1.x", versions)
			Expect(err).NotTo(HaveOccurred())
			Expect(vers).To(Equal([]string{"1.2.2", "1.2.3", "1.2.4", "1.3.2", "1.3.3", "1.3.4"}))
		})

		It("returns all versions in a patch line", func() {
			vers, err := bp.FindMatchingVersions("1.2.x", versions)
			Expect(err).NotTo(HaveOccurred())
			Expect(vers).To(Equal([]string{"1.2.2", "1.2.3", "1.2.4"}))
		})

		It("returns all versions less than the above", func() {
			vers, err := bp.FindMatchingVersions(">=1.2.0, <1.2.4", versions)
			Expect(err).NotTo(HaveOccurred())
			Expect(vers).To(Equal([]string{"1.2.2", "1.2.3"}))
		})

		It("returns all versions less than the above (without comma)", func() {
			vers, err := bp.FindMatchingVersions(">=1.2.0 <1.2.4", versions)
			Expect(err).NotTo(HaveOccurred())
			Expect(vers).To(Equal([]string{"1.2.2", "1.2.3"}))
		})

		It("returns all versions less or equal than the above (without comma)", func() {
			vers, err := bp.FindMatchingVersions(">1.2.2 <=1.2.4", versions)
			Expect(err).NotTo(HaveOccurred())
			Expect(vers).To(Equal([]string{"1.2.3", "1.2.4"}))
		})

		It("returns an error if nothing matches", func() {
			_, err := bp.FindMatchingVersions("1.4.x", versions)
			Expect(err).To(MatchError(fmt.Sprintf("no match found for 1.4.x in %v", versions)))
		})
	})
})
