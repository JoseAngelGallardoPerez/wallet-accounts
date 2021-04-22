package transfers_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestTransfers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Transfers Suite")
}
