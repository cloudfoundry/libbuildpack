package buildpack_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	ce "github.com/sesmith177/go-ce-test"
)

var _ = Describe("Manifest", func() {
	var (
		manifest     *ce.Manifest
		manifestFile string
		err          error
	)
	BeforeEach(func() {
		manifestFile = "fixtures/manifest.yml"
	})
	JustBeforeEach(func() {
		manifest, err = ce.NewManifest(manifestFile)
		Expect(err).To(BeNil())
	})

	Describe("NewManifest", func() {
		It("has a language", func() {
			Expect(manifest.Language).To(Equal("dotnet-core"))
		})
	})

	PDescribe("FetchDependency", func() {
		PContext("uncached", func() {})

		PContext("cached", func() {})
	})

	Describe("DefaultVersion", func() {
		Context("requested name exists (once)", func() {
			It("returns the default", func() {
				Expect(manifest.DefaultVersion("node")).To(Equal("6.9.4"))
			})
		})

		Context("requested name exists (twice)", func() {
			BeforeEach(func() { manifestFile = "fixtures/manifest_duplicate_default.yml" })
			It("returns an buildpack error", func() {
				_, err := manifest.DefaultVersion("bower")
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("found 2 default versions for bower"))
				Expect(err.(ce.Error).BuildpackError()).To(ContainSubstring("misconfigured for 'default_versions'"))
			})
		})
		Context("requested name does not exist", func() {
			It("returns an buildpack error", func() {
				_, err := manifest.DefaultVersion("notexist")
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("no default version for notexist"))
				Expect(err.(ce.Error).BuildpackError()).To(ContainSubstring("misconfigured for 'default_versions'"))
			})

		})

	})

})
