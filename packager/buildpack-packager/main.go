package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/cloudfoundry/libbuildpack/packager"
	"github.com/google/subcommands"
)

type summaryCmd struct {
}

func (*summaryCmd) Name() string             { return "summary" }
func (*summaryCmd) Synopsis() string         { return "Print out list of dependencies of this buildpack" }
func (*summaryCmd) SetFlags(f *flag.FlagSet) {}
func (*summaryCmd) Usage() string {
	return `summary:
  When run in a directory that is structured as a buildpack, prints a list of depedencies of that buildpack.
  (i.e. what would be downloaded to build a cached zipfile)
`
}
func (s *summaryCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	summary, err := packager.Summary(".")
	if err != nil {
		log.Printf("error reading dependencies from manifest: %v", err)
		return subcommands.ExitFailure
	}
	fmt.Println(summary)
	return subcommands.ExitSuccess
}

type buildCmd struct {
	cached   bool
	anyStack bool
	version  string
	cacheDir string
	stack    string
	profile  string
	exclude  string
	include  string
}

func (*buildCmd) Name() string     { return "build" }
func (*buildCmd) Synopsis() string { return "Create a buildpack zipfile from the current directory" }
func (*buildCmd) Usage() string {
	return `build -stack <stack>|-any-stack [-cached] [-version <version>] [-cachedir <path>]
      [-profile <profile>] [-exclude <dep1,dep2,...>] [-include <dep1,dep2,...>]:
  When run in a directory that is structured as a buildpack, creates a zip file.

  -profile  Name of a packaging profile defined in manifest.yml's
            packaging_profiles section. Profiles declare which dependencies
            to exclude from the cached zip.

  -exclude  Comma-separated list of dependency names to exclude, in addition
            to any exclusions implied by -profile. Names must exist in
            manifest.yml. Example: -exclude datadog-javaagent,newrelic

  -include  Comma-separated list of dependency names to force-include,
            overriding exclusions implied by -profile. Useful for starting
            from a restrictive profile and adding back a single dep.
            Example: -profile minimal -include jprofiler-profiler

`
}
func (b *buildCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&b.version, "version", "", "version to build as")
	f.BoolVar(&b.cached, "cached", false, "include dependencies")
	f.StringVar(&b.cacheDir, "cachedir", packager.CacheDir, "cache dir")
	f.StringVar(&b.stack, "stack", "", "stack to package buildpack for")
	f.BoolVar(&b.anyStack, "any-stack", false, "package buildpack for any stack")
	f.StringVar(&b.profile, "profile", "", "packaging profile defined in manifest.yml")
	f.StringVar(&b.exclude, "exclude", "", "comma-separated dependency names to exclude")
	f.StringVar(&b.include, "include", "", "comma-separated dependency names to include, overriding profile exclusions")
}
func (b *buildCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if b.stack == "" && !b.anyStack {
		log.Printf("error: must either specify a stack or pass -any-stack")
		return subcommands.ExitFailure
	}
	if b.stack != "" && b.anyStack {
		log.Printf("error: cannot specify a stack AND pass -any-stack")
		return subcommands.ExitFailure
	}
	if b.version == "" {
		v, err := os.ReadFile("VERSION")
		if err != nil {
			log.Printf("error: Could not read VERSION file: %v", err)
			return subcommands.ExitFailure
		}
		b.version = strings.TrimSpace(string(v))
	}

	parseCSV := func(s string) []string {
		var out []string
		for _, name := range strings.Split(s, ",") {
			name = strings.TrimSpace(name)
			if name != "" {
				out = append(out, name)
			}
		}
		return out
	}

	opts := packager.PackageOptions{
		Profile: b.profile,
		Exclude: parseCSV(b.exclude),
		Include: parseCSV(b.include),
	}

	zipFile, err := packager.PackageWithOptions(".", b.cacheDir, b.version, b.stack, b.cached, opts)
	if err != nil {
		log.Printf("error while creating zipfile: %v", err)
		return subcommands.ExitFailure
	}

	buildpackType := "uncached"
	if b.cached {
		buildpackType = "cached"
	}

	stat, err := os.Stat(zipFile)
	if err != nil {
		log.Printf("error while stating zipfile: %v", err)
		return subcommands.ExitFailure
	}

	fmt.Printf("%s buildpack created and saved as %s with a size of %dMB\n", buildpackType, zipFile, stat.Size()/1024/1024)
	return subcommands.ExitSuccess
}

func main() {
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(&summaryCmd{}, "Custom")
	subcommands.Register(&buildCmd{}, "Custom")

	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
