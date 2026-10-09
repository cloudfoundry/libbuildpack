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

		It("distinguishes 21.0.12 from 21.0.12.0 — 4-part wins regardless of input order", func() {
			for _, versions := range [][]string{
				{"21.0.12.0", "21.0.12"},
				{"21.0.12", "21.0.12.0"},
			} {
				ver, err := bp.FindMatchingVersion("21.x", versions)
				Expect(err).NotTo(HaveOccurred())
				Expect(ver).To(Equal("21.0.12.0"))
			}
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

	Describe("FindMatchingVersion with JEP 322 versions", func() {
		It("orders 21.0.12 < 21.0.12.1 < 21.0.12.1.1 regardless of input order", func() {
			for _, versions := range [][]string{
				{"21.0.12", "21.0.12.1", "21.0.12.1.1"},
				{"21.0.12.1.1", "21.0.12.1", "21.0.12"},
				{"21.0.12.1", "21.0.12.1.1", "21.0.12"},
			} {
				vers, err := bp.FindMatchingVersions("21.x", versions)
				Expect(err).NotTo(HaveOccurred())
				Expect(vers).To(Equal([]string{"21.0.12", "21.0.12.1", "21.0.12.1.1"}))
			}
		})

		It("orders by numeric build number", func() {
			for _, versions := range [][]string{
				{"21.0.12.1+1", "21.0.12.1+2", "21.0.12.1+10"},
				{"21.0.12.1+10", "21.0.12.1+2", "21.0.12.1+1"},
			} {
				vers, err := bp.FindMatchingVersions("21.x", versions)
				Expect(err).NotTo(HaveOccurred())
				Expect(vers).To(Equal([]string{"21.0.12.1+1", "21.0.12.1+2", "21.0.12.1+10"}))
			}
		})

		It("orders 3-part versions by numeric build number", func() {
			vers, err := bp.FindMatchingVersions("21.x", []string{"21.0.12+10", "21.0.12+9"})
			Expect(err).NotTo(HaveOccurred())
			Expect(vers).To(Equal([]string{"21.0.12+9", "21.0.12+10"}))
		})

		It("resolves the newest Liberica-style version in a mixed list", func() {
			versions := []string{"17.0.20+10", "17.0.20.1+2", "21.0.12+10", "21.0.12.1+1", "21.0.12.1+2", "25.0.4+9", "25.0.4.1+1"}
			for constraint, expected := range map[string]string{
				"17.x":   "17.0.20.1+2",
				"21.x":   "21.0.12.1+2",
				"21.0.x": "21.0.12.1+2",
				"21.*":   "21.0.12.1+2",
				"25.x":   "25.0.4.1+1",
			} {
				ver, err := bp.FindMatchingVersion(constraint, versions)
				Expect(err).NotTo(HaveOccurred())
				Expect(ver).To(Equal(expected), constraint)
			}
		})

		It("supports more than four numeric fields", func() {
			ver, err := bp.FindMatchingVersion("21.x", []string{"21.0.10", "21.0.10.0.1", "21.0.9.1"})
			Expect(err).NotTo(HaveOccurred())
			Expect(ver).To(Equal("21.0.10.0.1"))
		})

		It("handles Java 8 versions", func() {
			ver, err := bp.FindMatchingVersion("8.x", []string{"8.0.492+10", "8.0.504+7", "11.0.32.1+1"})
			Expect(err).NotTo(HaveOccurred())
			Expect(ver).To(Equal("8.0.504+7"))
		})

		It("matches an exact X.Y.Z.N+B version", func() {
			vers, err := bp.FindMatchingVersions("21.0.12.1+1", []string{"21.0.12+10", "21.0.12.1+1", "21.0.12.1+2"})
			Expect(err).NotTo(HaveOccurred())
			Expect(vers).To(Equal([]string{"21.0.12.1+1"}))
		})

		It("matches an exact X.Y.Z+B version without matching newer patch releases", func() {
			vers, err := bp.FindMatchingVersions("21.0.12+10", []string{"21.0.12+9", "21.0.12+10", "21.0.12.1+1"})
			Expect(err).NotTo(HaveOccurred())
			Expect(vers).To(Equal([]string{"21.0.12+10"}))
		})

		It("matches all builds of an exact X.Y.Z.N version without build number", func() {
			ver, err := bp.FindMatchingVersion("21.0.12.1", []string{"21.0.12+10", "21.0.12.1+2", "21.0.12.1+1"})
			Expect(err).NotTo(HaveOccurred())
			Expect(ver).To(Equal("21.0.12.1+2"))
		})

		It("orders numeric builds above non-numeric build metadata regardless of input order", func() {
			for _, versions := range [][]string{
				{"21.0.12+10", "21.0.12+foo", "21.0.12+9"},
				{"21.0.12+9", "21.0.12+foo", "21.0.12+10"},
				{"21.0.12+foo", "21.0.12+10", "21.0.12+9"},
				{"21.0.12+10", "21.0.12+9", "21.0.12+foo"},
			} {
				vers, err := bp.FindMatchingVersions("21.x", versions)
				Expect(err).NotTo(HaveOccurred())
				Expect(vers).To(Equal([]string{"21.0.12+foo", "21.0.12+9", "21.0.12+10"}))
			}
		})

		It("requires exact constraints to have the same number of fields", func() {
			vers, err := bp.FindMatchingVersions("21.0.12.0", []string{"21.0.12", "21.0.12.0", "21.0.12.0.0"})
			Expect(err).NotTo(HaveOccurred())
			Expect(vers).To(Equal([]string{"21.0.12.0"}))

			vers, err = bp.FindMatchingVersions("21.0.12+10", []string{"21.0.12.0+10", "21.0.12+10"})
			Expect(err).NotTo(HaveOccurred())
			Expect(vers).To(Equal([]string{"21.0.12+10"}))
		})

		It("returns error for an exact X.Y.Z.N+B version when the build is absent", func() {
			_, err := bp.FindMatchingVersion("21.0.12.1+3", []string{"21.0.12.1+1", "21.0.12.1+2"})
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
