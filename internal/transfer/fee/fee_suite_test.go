package fee_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestFee(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Fee Suite")
}
