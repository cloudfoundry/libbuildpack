package shims

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Finalizer struct {
	V2AppDir   string
	V3AppDir   string
	DepsIndex  string
	ProfileDir string
}

func (f *Finalizer) Finalize() error {

	if err := os.RemoveAll(f.V2AppDir); err != nil {
		return err
	}

	if err := os.Rename(f.V3AppDir, f.V2AppDir); err != nil {
		return err
	}

	profileContents := fmt.Sprintf(
		`export PACK_STACK_ID="org.cloudfoundry.stacks.%s"
export PACK_LAYERS_DIR="$DEPS_DIR"
export PACK_APP_DIR="$HOME"
exec $DEPS_DIR/v3-launcher "$2"
`,
		os.Getenv("CF_STACK"))

	return ioutil.WriteFile(filepath.Join(f.ProfileDir, "0_shim.sh"), []byte(profileContents), 0666)
}
