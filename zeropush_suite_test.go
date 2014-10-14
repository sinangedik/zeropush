package zeropush_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestZeropush(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Zeropush Suite")
}
