package limitserver_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestLimitserver(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Limitserver Suite")
}
