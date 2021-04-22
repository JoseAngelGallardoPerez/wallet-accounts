package limit_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestLimit(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Limit Suite")
}
