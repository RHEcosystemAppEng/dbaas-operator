package reconcilers

import (
	"os"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("FetchImageAndVersion", func() {
	Context("Missing env var", func() {
		It("invalid, should return empty value", func() {
			os.Unsetenv("FOO")
			Expect(fetchEnvValue("FOO")).To(BeEmpty())
		})
		It("valid, should return default values from embedded file - config/default/manager-env-images.yaml", func() {
			os.Unsetenv(dbaasDynamicPluginVersion)
			valArray := strings.Split(fetchEnvValue(dbaasDynamicPluginVersion), ":")
			Expect(valArray[0]).To(Equal(dbaasDynamicPluginName))
		})
	})

	Context("Existing env var", func() {
		It("should return set value", func() {
			imageTest := "test-image@sha256:fds45ds21kl"
			os.Setenv(dbaasDynamicPluginImg, imageTest)
			Expect(fetchEnvValue(dbaasDynamicPluginImg)).To(Equal(imageTest))
		})
	})
})

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "FetchEnvValue Suite")
}
