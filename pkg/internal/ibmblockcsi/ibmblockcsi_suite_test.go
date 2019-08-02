package ibmblockcsi_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestIbmblockcsi(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ibmblockcsi Suite")
}
