package tests

import (
	"testing"

	. "github.com/onsi/ginkgo/v2" //nolint: revive
	. "github.com/onsi/gomega"    //nolint: revive
)

func TestSimda(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SIMDA GRPC API Suite")
}
