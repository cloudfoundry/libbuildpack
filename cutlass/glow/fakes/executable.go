package fakes

import (
	"sync"

	"github.com/cloudfoundry/libbuildpack/cutlass/execution"
)

type Executable struct {
	ExecuteCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			Options execution.Options
			Args    []string
		}
		Returns struct {
			Stdout string
			Stderr string
			Err    error
		}
		Stub func(execution.Options, ...string) (string, string, error)
	}
}

func (f *Executable) Execute(param1 execution.Options, param2 ...string) (string, string, error) {
	f.ExecuteCall.Lock()
	defer f.ExecuteCall.Unlock()
	f.ExecuteCall.CallCount++
	f.ExecuteCall.Receives.Options = param1
	f.ExecuteCall.Receives.Args = param2
	if f.ExecuteCall.Stub != nil {
		return f.ExecuteCall.Stub(param1, param2...)
	}
	return f.ExecuteCall.Returns.Stdout, f.ExecuteCall.Returns.Stderr, f.ExecuteCall.Returns.Err
}
