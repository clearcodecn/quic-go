package http09

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestHttp09(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "HTTP/0.9 Suite")
}
