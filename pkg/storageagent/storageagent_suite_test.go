package storageagent

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestStorageAgent(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Storageagent Suite")
}
