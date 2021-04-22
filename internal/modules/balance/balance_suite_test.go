package balance_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestBalance(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Balance Suite")
}
