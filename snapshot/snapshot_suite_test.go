package snapshot_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSnapshot(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Snapshot Suite")
}
