package iscsi

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestIscsi(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Iscsi Suite")
}
