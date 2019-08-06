package fakeserver_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestFakeserver(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Fakeserver Suite")
}
