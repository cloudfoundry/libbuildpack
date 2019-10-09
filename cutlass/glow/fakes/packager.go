package fakes

import (
	"sync"

	"github.com/cloudfoundry/libbuildpack/cutlass/glow"
)

type Packager struct {
	PackageCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			Dir     string
			Stack   string
			Options glow.PackageOptions
		}
		Returns struct {
			Stdout string
			Stderr string
			Err    error
		}
		Stub func(string, string, glow.PackageOptions) (string, string, error)
	}
}

func (f *Packager) Package(param1 string, param2 string, param3 glow.PackageOptions) (string, string, error) {
	f.PackageCall.Lock()
	defer f.PackageCall.Unlock()
	f.PackageCall.CallCount++
	f.PackageCall.Receives.Dir = param1
	f.PackageCall.Receives.Stack = param2
	f.PackageCall.Receives.Options = param3
	if f.PackageCall.Stub != nil {
		return f.PackageCall.Stub(param1, param2, param3)
	}
	return f.PackageCall.Returns.Stdout, f.PackageCall.Returns.Stderr, f.PackageCall.Returns.Err
}
