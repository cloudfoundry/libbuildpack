package packager_test

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudfoundry/libbuildpack"
	"github.com/cloudfoundry/libbuildpack/packager"
	httpmock "github.com/jarcoal/httpmock"
	yaml "gopkg.in/yaml.v2"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// depNamesInManifest returns the dependency names listed in the manifest.yml
// that is embedded in the given zip file.
func depNamesInManifest(zipFile string) ([]string, error) {
	manifestYml, err := ZipContents(zipFile, "manifest.yml")
	if err != nil {
		return nil, err
	}
	var m packager.Manifest
	if err := yaml.Unmarshal([]byte(manifestYml), &m); err != nil {
		return nil, err
	}
	names := make([]string, 0, len(m.Dependencies))
	for _, d := range m.Dependencies {
		names = append(names, d.Name)
	}
	return names, nil
}

var _ = Describe("PackageWithOptions", func() {
	var (
		buildpackDir string
		version      string
		cacheDir     string
		stack        string
		zipFile      string
		err          error
	)

	BeforeEach(func() {
		stack = "cflinuxfs4"
		buildpackDir = "./fixtures/with_profiles"
		cacheDir, err = os.MkdirTemp("", "packager-cachedir")
		Expect(err).To(BeNil())
		version = fmt.Sprintf("1.0.0.%s", time.Now().Format("20060102150405"))
		httpmock.Reset()
	})

	AfterEach(func() {
		os.Remove(zipFile)
		os.RemoveAll(cacheDir)
	})

	// patchedFixtureDir returns (tmpDir, fixtureAbs) — a temp copy of
	// with_profiles where PLACEHOLDER_* URIs are replaced with real file://
	// paths so the packager can copy them during cached builds.
	patchedFixtureDir := func() (string, string) {
		GinkgoHelper()
		fixtureAbs, err := filepath.Abs("./fixtures/with_profiles")
		Expect(err).NotTo(HaveOccurred())

		tmpDir, err := os.MkdirTemp("", "bp_fixture_cached")
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(os.RemoveAll, tmpDir)

		Expect(libbuildpack.CopyDirectory(fixtureAbs, tmpDir)).To(Succeed())

		manifestPath := filepath.Join(tmpDir, "manifest.yml")
		raw, err := os.ReadFile(manifestPath)
		Expect(err).NotTo(HaveOccurred())

		patched := string(raw)
		patched = strings.ReplaceAll(patched, "file://PLACEHOLDER_CORE", "file://"+filepath.Join(fixtureAbs, "core.txt"))
		patched = strings.ReplaceAll(patched, "file://PLACEHOLDER_AGENT", "file://"+filepath.Join(fixtureAbs, "agent.txt"))
		patched = strings.ReplaceAll(patched, "file://PLACEHOLDER_PROFILER", "file://"+filepath.Join(fixtureAbs, "profiler.txt"))

		Expect(os.WriteFile(manifestPath, []byte(patched), 0644)).To(Succeed())
		return tmpDir, fixtureAbs
	}

	Context("no profile, no exclude, no include (zero opts)", func() {
		JustBeforeEach(func() {
			zipFile, err = packager.PackageWithOptions(buildpackDir, cacheDir, version, stack, false, packager.PackageOptions{})
			Expect(err).To(BeNil())
		})

		It("bundles all stack-matching dependencies", func() {
			names, err := depNamesInManifest(zipFile)
			Expect(err).To(BeNil())
			Expect(names).To(ConsistOf("core-dep", "agent-dep", "profiler-dep"))
		})

		It("produces the standard filename with no profile suffix", func() {
			dir, _ := filepath.Abs(buildpackDir)
			Expect(zipFile).To(Equal(filepath.Join(dir, fmt.Sprintf("ruby_buildpack-cflinuxfs4-v%s.zip", version))))
		})
	})

	Context("explicit --exclude of one dependency", func() {
		It("omits the excluded dependency from the manifest", func() {
			dir, _ := patchedFixtureDir()
			zipFile, err = packager.PackageWithOptions(dir, cacheDir, version, stack, true, packager.PackageOptions{
				Exclude: []string{"agent-dep"},
			})
			Expect(err).To(BeNil())
			names, err := depNamesInManifest(zipFile)
			Expect(err).To(BeNil())
			Expect(names).To(ConsistOf("core-dep", "profiler-dep"))
			Expect(names).NotTo(ContainElement("agent-dep"))
		})
	})

	Context("named profile that excludes two dependencies", func() {
		It("omits all profile-excluded dependencies", func() {
			dir, _ := patchedFixtureDir()
			zipFile, err = packager.PackageWithOptions(dir, cacheDir, version, stack, true, packager.PackageOptions{
				Profile: "minimal",
			})
			Expect(err).To(BeNil())
			names, err := depNamesInManifest(zipFile)
			Expect(err).To(BeNil())
			Expect(names).To(ConsistOf("core-dep"))
			Expect(names).NotTo(ContainElement("agent-dep"))
			Expect(names).NotTo(ContainElement("profiler-dep"))
		})
	})

	Context("profile combined with extra --exclude", func() {
		It("applies union of profile and explicit excludes", func() {
			dir, _ := patchedFixtureDir()
			// no-profiler only excludes profiler-dep; adding --exclude agent-dep unions both
			zipFile, err = packager.PackageWithOptions(dir, cacheDir, version, stack, true, packager.PackageOptions{
				Profile: "no-profiler",
				Exclude: []string{"agent-dep"},
			})
			Expect(err).To(BeNil())
			names, err := depNamesInManifest(zipFile)
			Expect(err).To(BeNil())
			Expect(names).To(ConsistOf("core-dep"))
		})
	})

	Context("profile with --include restoring an excluded dependency", func() {
		It("restores the included dependency while keeping other exclusions", func() {
			dir, _ := patchedFixtureDir()
			// minimal excludes both agent-dep and profiler-dep; --include restores profiler-dep
			zipFile, err = packager.PackageWithOptions(dir, cacheDir, version, stack, true, packager.PackageOptions{
				Profile: "minimal",
				Include: []string{"profiler-dep"},
			})
			Expect(err).To(BeNil())
			names, err := depNamesInManifest(zipFile)
			Expect(err).To(BeNil())
			Expect(names).To(ConsistOf("core-dep", "profiler-dep"))
			Expect(names).NotTo(ContainElement("agent-dep"))
		})
	})

	Context("profile with --exclude and --include combined", func() {
		It("applies exclude union then include override", func() {
			dir, _ := patchedFixtureDir()
			// no-profiler excludes profiler-dep; we also exclude agent-dep but restore profiler-dep
			zipFile, err = packager.PackageWithOptions(dir, cacheDir, version, stack, true, packager.PackageOptions{
				Profile: "no-profiler",
				Exclude: []string{"agent-dep"},
				Include: []string{"profiler-dep"},
			})
			Expect(err).To(BeNil())
			names, err := depNamesInManifest(zipFile)
			Expect(err).To(BeNil())
			Expect(names).To(ConsistOf("core-dep", "profiler-dep"))
		})
	})

	Context("--include without --profile", func() {
		It("returns an error", func() {
			zipFile, err = packager.PackageWithOptions(buildpackDir, cacheDir, version, stack, true, packager.PackageOptions{
				Include: []string{"core-dep"},
			})
			Expect(err).To(MatchError(ContainSubstring("--include requires --profile")))
		})
	})

	Context("uncached buildpack rejects profile/exclude/include flags", func() {
		It("returns an error when --profile is used on an uncached build", func() {
			zipFile, err = packager.PackageWithOptions(buildpackDir, cacheDir, version, stack, false, packager.PackageOptions{
				Profile: "minimal",
			})
			Expect(err).To(MatchError(ContainSubstring("only valid for cached buildpacks")))
		})

		It("returns an error when --exclude is used on an uncached build", func() {
			zipFile, err = packager.PackageWithOptions(buildpackDir, cacheDir, version, stack, false, packager.PackageOptions{
				Exclude: []string{"agent-dep"},
			})
			Expect(err).To(MatchError(ContainSubstring("only valid for cached buildpacks")))
		})

		It("returns an error when --include is used on an uncached build", func() {
			zipFile, err = packager.PackageWithOptions(buildpackDir, cacheDir, version, stack, false, packager.PackageOptions{
				Profile: "minimal",
				Include: []string{"profiler-dep"},
			})
			Expect(err).To(MatchError(ContainSubstring("only valid for cached buildpacks")))
		})
	})

	Context("manifest validation errors (cached=true)", func() {
		It("unknown profile name returns an error", func() {
			dir, _ := patchedFixtureDir()
			zipFile, err = packager.PackageWithOptions(dir, cacheDir, version, stack, true, packager.PackageOptions{
				Profile: "does-not-exist",
			})
			Expect(err).To(MatchError(ContainSubstring(`packaging profile "does-not-exist" not found in manifest`)))
		})

		It("--exclude with unknown dependency name returns an error", func() {
			dir, _ := patchedFixtureDir()
			zipFile, err = packager.PackageWithOptions(dir, cacheDir, version, stack, true, packager.PackageOptions{
				Exclude: []string{"no-such-dep"},
			})
			Expect(err).To(MatchError(ContainSubstring(`dependency "no-such-dep" not found in manifest`)))
		})

		It("--include with unknown dependency name returns an error", func() {
			dir, _ := patchedFixtureDir()
			zipFile, err = packager.PackageWithOptions(dir, cacheDir, version, stack, true, packager.PackageOptions{
				Profile: "minimal",
				Include: []string{"no-such-dep"},
			})
			Expect(err).To(MatchError(ContainSubstring(`dependency "no-such-dep" not found in manifest`)))
		})

		It("profile with a bad exclude name returns an error referencing the profile", func() {
			dir, _ := patchedFixtureDir()
			// Inject a profile whose exclude list names a non-existent dep.
			badManifest := filepath.Join(dir, "manifest.yml")
			raw, readErr := os.ReadFile(badManifest)
			Expect(readErr).To(BeNil())
			patched := strings.Replace(string(raw),
				"packaging_profiles:",
				"packaging_profiles:\n  bad-exclude:\n    description: profile with bogus dep\n    exclude:\n    - no-such-dep",
				1)
			Expect(os.WriteFile(badManifest, []byte(patched), 0644)).To(Succeed())

			zipFile, err = packager.PackageWithOptions(dir, cacheDir, version, stack, true, packager.PackageOptions{
				Profile: "bad-exclude",
			})
			Expect(err).To(MatchError(ContainSubstring(`profile "bad-exclude" references unknown dependency "no-such-dep"`)))
		})

		It("invalid profile name (contains spaces) returns an error", func() {
			dir, _ := patchedFixtureDir()
			zipFile, err = packager.PackageWithOptions(dir, cacheDir, version, stack, true, packager.PackageOptions{
				Profile: "my profile",
			})
			Expect(err).To(MatchError(ContainSubstring(`profile name "my profile" is invalid`)))
		})

		It("invalid profile name (contains slash) returns an error", func() {
			dir, _ := patchedFixtureDir()
			zipFile, err = packager.PackageWithOptions(dir, cacheDir, version, stack, true, packager.PackageOptions{
				Profile: "../../etc/passwd",
			})
			Expect(err).To(MatchError(ContainSubstring(`profile name "../../etc/passwd" is invalid`)))
		})

		It("--include of dep not excluded by profile or --exclude returns an error", func() {
			dir, _ := patchedFixtureDir()
			// no-profiler excludes profiler-dep; core-dep is never excluded → hard error
			zipFile, err = packager.PackageWithOptions(dir, cacheDir, version, stack, true, packager.PackageOptions{
				Profile: "no-profiler",
				Include: []string{"core-dep"},
			})
			Expect(err).To(MatchError(ContainSubstring(`--include "core-dep" has no effect`)))
		})
	})

	Context("backward compat: Package() delegates to PackageWithOptions with zero opts", func() {
		JustBeforeEach(func() {
			zipFile, err = packager.Package(buildpackDir, cacheDir, version, stack, false)
			Expect(err).To(BeNil())
		})

		It("bundles all stack-matching dependencies", func() {
			names, err := depNamesInManifest(zipFile)
			Expect(err).To(BeNil())
			Expect(names).To(ConsistOf("core-dep", "agent-dep", "profiler-dep"))
		})

		It("produces the standard filename with no profile suffix", func() {
			dir, _ := filepath.Abs(buildpackDir)
			Expect(zipFile).To(Equal(filepath.Join(dir, fmt.Sprintf("ruby_buildpack-cflinuxfs4-v%s.zip", version))))
		})
	})

	Context("zip filename variants", func() {
		// Opts-bearing tests need cached=true + real file:// URIs.
		// The zero-opts test stays uncached (no downloads needed).

		It("no opts → no profile suffix", func() {
			zipFile, err = packager.PackageWithOptions(buildpackDir, cacheDir, version, stack, false, packager.PackageOptions{})
			Expect(err).To(BeNil())
			dir, _ := filepath.Abs(buildpackDir)
			Expect(zipFile).To(Equal(filepath.Join(dir, fmt.Sprintf("ruby_buildpack-cflinuxfs4-v%s.zip", version))))
		})

		It("profile only → -<profile>", func() {
			dir, _ := patchedFixtureDir()
			zipFile, err = packager.PackageWithOptions(dir, cacheDir, version, stack, true, packager.PackageOptions{Profile: "minimal"})
			Expect(err).To(BeNil())
			Expect(zipFile).To(Equal(filepath.Join(dir, fmt.Sprintf("ruby_buildpack-cached-minimal-cflinuxfs4-v%s.zip", version))))
		})

		It("profile + exclude → -<profile>+custom", func() {
			dir, _ := patchedFixtureDir()
			zipFile, err = packager.PackageWithOptions(dir, cacheDir, version, stack, true, packager.PackageOptions{
				Profile: "no-profiler", Exclude: []string{"agent-dep"},
			})
			Expect(err).To(BeNil())
			Expect(zipFile).To(Equal(filepath.Join(dir, fmt.Sprintf("ruby_buildpack-cached-no-profiler+custom-cflinuxfs4-v%s.zip", version))))
		})

		It("profile + include that overrides a profile exclusion → -<profile>+custom", func() {
			dir, _ := patchedFixtureDir()
			// minimal excludes agent-dep and profiler-dep; restoring profiler-dep is an effective override
			zipFile, err = packager.PackageWithOptions(dir, cacheDir, version, stack, true, packager.PackageOptions{
				Profile: "minimal", Include: []string{"profiler-dep"},
			})
			Expect(err).To(BeNil())
			Expect(zipFile).To(Equal(filepath.Join(dir, fmt.Sprintf("ruby_buildpack-cached-minimal+custom-cflinuxfs4-v%s.zip", version))))
		})

		It("profile + include of dep NOT in profile exclusions → hard error", func() {
			dir, _ := patchedFixtureDir()
			// no-profiler only excludes profiler-dep; --include core-dep was never excluded → hard error
			zipFile, err = packager.PackageWithOptions(dir, cacheDir, version, stack, true, packager.PackageOptions{
				Profile: "no-profiler", Include: []string{"core-dep"},
			})
			Expect(err).To(MatchError(ContainSubstring(`--include "core-dep" has no effect`)))
		})

		It("exclude only (no profile) → -custom", func() {
			dir, _ := patchedFixtureDir()
			zipFile, err = packager.PackageWithOptions(dir, cacheDir, version, stack, true, packager.PackageOptions{
				Exclude: []string{"agent-dep"},
			})
			Expect(err).To(BeNil())
			Expect(zipFile).To(Equal(filepath.Join(dir, fmt.Sprintf("ruby_buildpack-cached-custom-cflinuxfs4-v%s.zip", version))))
		})
	})

	Context("--exclude multiple dependencies at once", func() {
		It("omits all listed deps from the packaged manifest", func() {
			dir, _ := patchedFixtureDir()
			zipFile, err = packager.PackageWithOptions(dir, cacheDir, version, stack, true, packager.PackageOptions{
				Exclude: []string{"agent-dep", "profiler-dep"},
			})
			Expect(err).To(BeNil())
			names, err := depNamesInManifest(zipFile)
			Expect(err).To(BeNil())
			Expect(names).To(ConsistOf("core-dep"))
		})
	})

	Context("cached=true with exclusion (excluded dep not downloaded, not in zip)", func() {
		It("excluded dependency is not present as a binary file in the zip", func() {
			dir, fixtureAbs := patchedFixtureDir()
			agentURI := "file://" + filepath.Join(fixtureAbs, "agent.txt")

			zipFile, err = packager.PackageWithOptions(dir, cacheDir, version, stack, true, packager.PackageOptions{
				Exclude: []string{"agent-dep"},
			})
			Expect(err).To(BeNil())

			// The agent binary path inside the zip is keyed by the MD5 of its URI.
			agentZipPath := fmt.Sprintf("dependencies/%x/agent.txt", md5.Sum([]byte(agentURI)))
			_, lookupErr := ZipContents(zipFile, agentZipPath)
			Expect(lookupErr).To(MatchError(ContainSubstring("not found in")))
		})

		It("non-excluded dependencies are still downloaded and present in the zip", func() {
			dir, _ := patchedFixtureDir()
			zipFile, err = packager.PackageWithOptions(dir, cacheDir, version, stack, true, packager.PackageOptions{
				Exclude: []string{"agent-dep"},
			})
			Expect(err).To(BeNil())

			names, err := depNamesInManifest(zipFile)
			Expect(err).To(BeNil())
			Expect(names).To(ConsistOf("core-dep", "profiler-dep"))
		})

		It("profile + include: restored dep is downloaded and present in the manifest", func() {
			dir, _ := patchedFixtureDir()
			// minimal excludes both agent-dep and profiler-dep; --include restores profiler-dep
			zipFile, err = packager.PackageWithOptions(dir, cacheDir, version, stack, true, packager.PackageOptions{
				Profile: "minimal",
				Include: []string{"profiler-dep"},
			})
			Expect(err).To(BeNil())

			names, err := depNamesInManifest(zipFile)
			Expect(err).To(BeNil())
			Expect(names).To(ConsistOf("core-dep", "profiler-dep"))
			Expect(names).NotTo(ContainElement("agent-dep"))
		})
	})
})

var _ = Describe("Packager", func() {
	var (
		buildpackDir string
		version      string
		cacheDir     string
		stack        string
		err          error
	)

	BeforeEach(func() {
		stack = "cflinuxfs2"
		buildpackDir = "./fixtures/good"
		cacheDir, err = os.MkdirTemp("", "packager-cachedir")
		Expect(err).To(BeNil())
		version = fmt.Sprintf("1.23.45.%s", time.Now().Format("20060102150405"))

		httpmock.Reset()
	})

	Describe("DownloadFromURI", func() {
		var (
			destFile string
			destDir  string
		)

		BeforeEach(func() {
			var err error
			destDir, err = os.MkdirTemp("", "packager-download")
			Expect(err).To(BeNil())
			DeferCleanup(os.RemoveAll, destDir)
			destFile = filepath.Join(destDir, "some_dep_1.2.3+4_linux.tgz")
		})

		Context("when the server returns a non-2xx status", func() {
			It("returns an error containing the URI and does not leave a file on disk", func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					http.Error(w, "Forbidden", http.StatusForbidden)
				}))
				defer server.Close()

				uri := server.URL + "/dependencies/dep/dep_1.2.3+4_linux.tgz"
				err := packager.DownloadFromURI(uri, destFile)
				Expect(err).To(MatchError(ContainSubstring("could not download")))
				Expect(err).To(MatchError(ContainSubstring(uri)))
				Expect(err).To(MatchError(ContainSubstring("403")))

				_, statErr := os.Stat(destFile)
				Expect(os.IsNotExist(statErr)).To(BeTrue(), "stale file should not be left on disk after a failed download")
			})
		})

		Context("when the download succeeds", func() {
			It("writes the file to disk", func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("binary content"))
				}))
				defer server.Close()

				uri := server.URL + "/dependencies/dep/dep_1.2.3_linux.tgz"
				err := packager.DownloadFromURI(uri, destFile)
				Expect(err).To(BeNil())

				_, statErr := os.Stat(destFile)
				Expect(statErr).To(BeNil(), "file should exist after successful download")
			})
		})
	})

	Describe("Package", func() {
		var zipFile string
		var cached bool
		BeforeEach(func() {
			DeferCleanup(func() { os.Remove(zipFile) })
		})

		AssertStack := func() {
			var manifest *packager.Manifest
			Context("stack specified and matches any dependency in manifest.yml", func() {
				BeforeEach(func() { stack = "cflinuxfs2" })
				JustBeforeEach(func() {
					manifestYml, err := ZipContents(zipFile, "manifest.yml")
					Expect(err).To(BeNil())
					manifest = &packager.Manifest{}
					Expect(yaml.Unmarshal([]byte(manifestYml), manifest)).To(Succeed())
				})

				It("removes dependencies for other stacks from the manifest", func() {
					Expect(len(manifest.Dependencies)).To(Equal(1))
					Expect(manifest.Dependencies[0].SHA256).To(Equal("b11329c3fd6dbe9dddcb8dd90f18a4bf441858a6b5bfaccae5f91e5c7d2b3596"))
				})

				It("removes cfstacks from the remaining dependencies", func() {
					Expect(manifest.Dependencies[0].Stacks).To(BeNil())
				})

				It("adds a top-level stack: key to the manifest", func() {
					Expect(manifest.Stack).To(Equal(stack))
				})
			})

			Context("empty stack specified", func() {
				BeforeEach(func() { stack = "" })
				JustBeforeEach(func() {
					manifestYml, err := ZipContents(zipFile, "manifest.yml")
					Expect(err).To(BeNil())
					manifest = &packager.Manifest{}
					Expect(yaml.Unmarshal([]byte(manifestYml), manifest)).To(Succeed())
				})

				It("includes dependencies for all stacks in the manifest", func() {
					Expect(len(manifest.Dependencies)).To(Equal(2))
				})

				It("does not add a top-level stack: key to the manifest", func() {
					Expect(manifest.Stack).To(Equal(""))
				})

				It("does not remove cf_stacks from dependencies", func() {
					Expect(manifest.Dependencies[0].Stacks).To(Equal([]string{"cflinuxfs2"}))
					Expect(manifest.Dependencies[1].Stacks).To(Equal([]string{"cflinuxfs3"}))
				})
			})
		}

		Context("uncached", func() {
			BeforeEach(func() { cached = false })
			JustBeforeEach(func() {
				var err error
				zipFile, err = packager.Package(buildpackDir, cacheDir, version, stack, cached)
				Expect(err).To(BeNil())
			})

			AssertStack()

			It("generates a zipfile with name", func() {
				dir, err := filepath.Abs(buildpackDir)
				Expect(err).To(BeNil())
				Expect(zipFile).To(Equal(filepath.Join(dir, fmt.Sprintf("ruby_buildpack-cflinuxfs2-v%s.zip", version))))
			})

			It("includes files listed in manifest.yml", func() {
				Expect(ZipContents(zipFile, "bin/filename")).To(Equal("awesome content"))
			})

			It("overrides VERSION", func() {
				Expect(ZipContents(zipFile, "VERSION")).To(Equal(version))
			})

			It("runs pre-package script", func() {
				Expect(ZipContents(zipFile, "hi.txt")).To(Equal("hi mom\n"))
			})

			It("does not include files not in list", func() {
				_, err := ZipContents(zipFile, "ignoredfile")
				Expect(err).To(MatchError(HavePrefix("ignoredfile not found in")))
			})

			It("does not include dependencies", func() {
				_, err := ZipContents(zipFile, "dependencies/d39cae561ec1f485d1a4a58304e87105/rfc2324.txt")
				Expect(err).To(MatchError(HavePrefix("dependencies/d39cae561ec1f485d1a4a58304e87105/rfc2324.txt not found in")))
			})

			It("does not set file on entries", func() {
				manifestYml, err := ZipContents(zipFile, "manifest.yml")
				Expect(err).To(BeNil())
				var m packager.Manifest
				Expect(yaml.Unmarshal([]byte(manifestYml), &m)).To(Succeed())
				Expect(m.Dependencies).ToNot(BeEmpty())
				Expect(m.Dependencies[0].File).To(Equal(""))
			})
		})

		Context("cached", func() {
			BeforeEach(func() { cached = true })
			JustBeforeEach(func() {
				var err error
				zipFile, err = packager.Package(buildpackDir, cacheDir, version, stack, cached)
				Expect(err).To(BeNil())
			})

			AssertStack()

			It("generates a zipfile with name", func() {
				dir, err := filepath.Abs(buildpackDir)
				Expect(err).To(BeNil())
				Expect(zipFile).To(Equal(filepath.Join(dir, fmt.Sprintf("ruby_buildpack-cached-cflinuxfs2-v%s.zip", version))))
			})

			It("includes files listed in manifest.yml", func() {
				Expect(ZipContents(zipFile, "bin/filename")).To(Equal("awesome content"))
			})

			It("overrides VERSION", func() {
				Expect(ZipContents(zipFile, "VERSION")).To(Equal(version))
			})

			It("runs pre-package script", func() {
				Expect(ZipContents(zipFile, "hi.txt")).To(Equal("hi mom\n"))
			})

			It("does not include files not in list", func() {
				_, err := ZipContents(zipFile, "ignoredfile")
				Expect(err).To(MatchError(HavePrefix("ignoredfile not found in")))
			})

			Describe("including dependencies", func() {
				Context("when a stack is specified", func() {
					It("includes ONLY dependencies for the specified stack", func() {
						Expect(ZipContents(zipFile, "dependencies/d39cae561ec1f485d1a4a58304e87105/rfc2324.txt")).To(ContainSubstring("Hyper Text Coffee Pot Control Protocol"))
						_, err := ZipContents(zipFile, "dependencies/ff1eb131521acf5bc95db59b2a2c29c0/rfc2549.txt")
						Expect(err).To(MatchError(HavePrefix("dependencies/ff1eb131521acf5bc95db59b2a2c29c0/rfc2549.txt not found in")))
					})
				})
				Context("when the empty stack is specified", func() {
					BeforeEach(func() { stack = "" })

					It("includes dependencies for ALL stacks if the empty stack is used", func() {
						Expect(ZipContents(zipFile, "dependencies/d39cae561ec1f485d1a4a58304e87105/rfc2324.txt")).To(ContainSubstring("Hyper Text Coffee Pot Control Protocol"))
						Expect(ZipContents(zipFile, "dependencies/ff1eb131521acf5bc95db59b2a2c29c0/rfc2549.txt")).To(ContainSubstring("IP over Avian Carriers with Quality of Service"))
					})
				})
			})

			It("sets file on entries", func() {
				manifestYml, err := ZipContents(zipFile, "manifest.yml")
				Expect(err).To(BeNil())
				var m packager.Manifest
				Expect(yaml.Unmarshal([]byte(manifestYml), &m)).To(Succeed())
				Expect(m.Dependencies).ToNot(BeEmpty())
				Expect(m.Dependencies[0].File).To(Equal("dependencies/d39cae561ec1f485d1a4a58304e87105/rfc2324.txt"))
			})

			Context("dependency uses file://", func() {
				var tempfile string
				BeforeEach(func() {
					var err error
					tempdir, err := os.MkdirTemp("", "bp_fixture")
					Expect(err).ToNot(HaveOccurred())
					Expect(libbuildpack.CopyDirectory(buildpackDir, tempdir)).To(Succeed())

					fh, err := os.CreateTemp("", "bp_dependency")
					Expect(err).ToNot(HaveOccurred())
					fh.WriteString("keaty")
					fh.Close()
					tempfile = fh.Name()

					manifestyml, err := os.ReadFile(filepath.Join(tempdir, "manifest.yml"))
					Expect(err).ToNot(HaveOccurred())
					manifestyml2 := string(manifestyml)
					manifestyml2 = strings.Replace(manifestyml2, "https://www.ietf.org/rfc/rfc2324.txt", "file://"+tempfile, -1)
					manifestyml2 = strings.Replace(manifestyml2, "b11329c3fd6dbe9dddcb8dd90f18a4bf441858a6b5bfaccae5f91e5c7d2b3596", "f909ee4c4bec3280bbbff6b41529479366ab10c602d8aed33e3a86f0a9c5db4e", -1)
					Expect(os.WriteFile(filepath.Join(tempdir, "manifest.yml"), []byte(manifestyml2), 0644)).To(Succeed())

					buildpackDir = tempdir
					DeferCleanup(os.RemoveAll, buildpackDir)
				})

				It("includes dependencies", func() {
					dest := filepath.Join("dependencies", fmt.Sprintf("%x", md5.Sum([]byte("file://"+tempfile))), filepath.Base(tempfile))
					Expect(ZipContents(zipFile, dest)).To(ContainSubstring("keaty"))
				})
			})
		})

		Context("manifest.yml was already packaged", func() {
			Context("setting specific stack", func() {
				BeforeEach(func() { stack = "cflinuxfs2" })
				It("returns an error", func() {
					zipFile, err = packager.Package("./fixtures/prepackaged", cacheDir, version, stack, cached)
					Expect(err).To(MatchError("Cannot package from already packaged buildpack manifest"))
				})
			})

			Context("setting any stack", func() {

				BeforeEach(func() { stack = "" })

				It("returns an error", func() {
					zipFile, err = packager.Package("./fixtures/prepackaged", cacheDir, version, stack, cached)
					Expect(err).To(MatchError("Cannot package from already packaged buildpack manifest"))
				})
			})
		})

		Context("manifest.yml has no dependencies", func() {
			BeforeEach(func() { stack = "cflinuxfs2" })

			It("allows stack when packaging", func() {
				zipFile, err = packager.Package("./fixtures/no_dependencies", cacheDir, version, stack, cached)
				Expect(err).To(BeNil())
			})
		})

		Context("stack is invalid", func() {
			Context("stack not found in any dependencies", func() {
				BeforeEach(func() { stack = "nonexistent-stack" })

				It("returns an error", func() {
					zipFile, err = packager.Package(buildpackDir, cacheDir, version, stack, cached)
					Expect(err).To(MatchError("Stack `nonexistent-stack` not found in manifest"))
				})
			})
			Context("stack not found in any default dependencies", func() {
				BeforeEach(func() {
					stack = "cflinuxfs3"
					buildpackDir = "./fixtures/missing_default_fs3"
				})

				It("returns an error", func() {
					zipFile, err = packager.Package(buildpackDir, cacheDir, version, stack, cached)
					Expect(err).To(MatchError("No matching default dependency `ruby` for stack `cflinuxfs3`"))
				})
			})
		})

		Context("when buildpack includes symlink to directory", func() {
			BeforeEach(func() {
				// this is actually a failing test....
				//if runtime.GOOS == "darwin" {
				//	Skip("io.Copy fails with symlinked directory on Darwin")
				//}
				cached = true
				buildpackDir = "./fixtures/symlink_dir"
			})
			JustBeforeEach(func() {
				var err error
				zipFile, err = packager.Package(buildpackDir, cacheDir, version, stack, cached)
				Expect(err).To(BeNil())
			})
			It("gets zipfile name", func() {
				Expect(zipFile).ToNot(BeEmpty())
			})
			It("generates a zipfile with name", func() {
				var cachedStr string
				if cached {
					cachedStr = "-cached"
				}
				dir, err := filepath.Abs(buildpackDir)
				Expect(err).To(BeNil())
				Expect(zipFile).To(Equal(filepath.Join(dir, fmt.Sprintf("ruby_buildpack%s-cflinuxfs2-v%s.zip", cachedStr, version))))
			})

			It("includes files listed in manifest.yml", func() {
				Expect(ZipContents(zipFile, "bin/filename")).To(Equal("awesome content"))
			})

			It("overrides VERSION", func() {
				Expect(ZipContents(zipFile, "VERSION")).To(Equal(version))
			})

			It("runs pre-package script", func() {
				Expect(ZipContents(zipFile, "hi.txt")).To(Equal("hi mom\n"))
			})

			It("does not include files not in list", func() {
				_, err := ZipContents(zipFile, "ignoredfile")
				Expect(err).To(MatchError(HavePrefix("ignoredfile not found in")))
			})
		})

		Context("cached dependency has wrong md5", func() {
			BeforeEach(func() {
				cached = true
				buildpackDir = "./fixtures/bad"
			})
			It("includes dependencies", func() {
				zipFile, err = packager.Package(buildpackDir, cacheDir, version, stack, cached)
				Expect(err).To(MatchError(ContainSubstring("dependency sha256 mismatch: expected sha256 fffffff, actual sha256 b11329c3fd6dbe9dddcb8dd90f18a4bf441858a6b5bfaccae5f91e5c7d2b3596")))
			})
		})

		Context("packaging with no stack", func() {
			BeforeEach(func() {
				cached = false
				stack = ""
			})

			JustBeforeEach(func() {
				var err error
				zipFile, err = packager.Package(buildpackDir, cacheDir, version, stack, cached)
				Expect(err).To(BeNil())
			})

			It("generates a zipfile with name", func() {
				dir, err := filepath.Abs(buildpackDir)
				Expect(err).To(BeNil())
				Expect(zipFile).To(Equal(filepath.Join(dir, fmt.Sprintf("ruby_buildpack-v%s.zip", version))))
			})
		})

		Context("packaging with missing included_files", func() {
			It("returns an error", func() {
				zipFile, err = packager.Package("./fixtures/missing_included_files", cacheDir, version, stack, cached)
				Expect(err).To(MatchError(MatchRegexp("failed to open included_file: .*/DOESNOTEXIST.txt")))
			})
		})
	})
})
