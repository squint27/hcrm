//go:build integration
// +build integration

package hcloud

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestHCloudIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "HCloud Integration Suite")
}
