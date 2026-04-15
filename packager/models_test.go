package packager_test

import (
	"sort"

	"github.com/cloudfoundry/libbuildpack/packager"
	httpmock "github.com/jarcoal/httpmock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	yaml "gopkg.in/yaml.v2"
)

var _ = Describe("Packager", func() {
	BeforeEach(func() {
		httpmock.Reset()
	})

	Describe("Sort Dependencies", func() {
		It("....", func() {
			deps := packager.Dependencies{
				{Name: "ruby", Version: "1.2.3"},
				{Name: "ruby", Version: "3.2.1"},
				{Name: "zesty", Version: "2.1.3"},
				{Name: "ruby", Version: "1.11.3"},
				{Name: "jruby", Version: "2.1.3"},
			}
			sort.Sort(deps)
			Expect(deps).To(Equal(packager.Dependencies{
				{Name: "jruby", Version: "2.1.3"},
				{Name: "ruby", Version: "1.2.3"},
				{Name: "ruby", Version: "1.11.3"},
				{Name: "ruby", Version: "3.2.1"},
				{Name: "zesty", Version: "2.1.3"},
			}))
		})
	})

	Describe("PackagingProfile YAML unmarshalling", func() {
		It("parses packaging_profiles from manifest YAML", func() {
			raw := `
language: ruby
dependencies: []
packaging_profiles:
  minimal:
    description: "JDKs only"
    exclude:
      - agent-dep
      - profiler-dep
  standard:
    description: "Core and OSS agents"
    exclude:
      - profiler-dep
`
			var m packager.Manifest
			Expect(yaml.Unmarshal([]byte(raw), &m)).To(Succeed())
			Expect(m.PackagingProfiles).To(HaveLen(2))

			minimal := m.PackagingProfiles["minimal"]
			Expect(minimal.Description).To(Equal("JDKs only"))
			Expect(minimal.Exclude).To(ConsistOf("agent-dep", "profiler-dep"))

			standard := m.PackagingProfiles["standard"]
			Expect(standard.Description).To(Equal("Core and OSS agents"))
			Expect(standard.Exclude).To(ConsistOf("profiler-dep"))
		})

		It("has nil PackagingProfiles when the section is absent", func() {
			raw := `language: ruby
dependencies: []
`
			var m packager.Manifest
			Expect(yaml.Unmarshal([]byte(raw), &m)).To(Succeed())
			Expect(m.PackagingProfiles).To(BeNil())
		})
	})
})
