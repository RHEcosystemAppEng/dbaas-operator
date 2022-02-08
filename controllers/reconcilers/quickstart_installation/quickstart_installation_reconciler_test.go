package quickstart_installation

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	consolev1 "github.com/openshift/api/console/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

var _ = Describe("ConsoleQuickStart installation reconciler", func() {
	Context("ConsoleQuickStart autogeneration", func() {
		It("should unmarshal quickstart CR's correctly", func() {

			Expect(QuickStarts).ShouldNot(BeEmpty())
			for qsName, qsBytes := range QuickStarts {
				qsExpected := &consolev1.ConsoleQuickStart{
					ObjectMeta: metav1.ObjectMeta{
						Name: qsName,
						Annotations: map[string]string{
							"categories": "Database management",
						},
					},
				}
				quickStartFromFile := &consolev1.ConsoleQuickStart{}
				err := yaml.Unmarshal(qsBytes, quickStartFromFile)
				Expect(err).NotTo(HaveOccurred())
				Expect(quickStartFromFile.Name).Should(Equal(qsExpected.Name))
				Expect(quickStartFromFile.Annotations).Should(Equal(qsExpected.Annotations))
				Expect(quickStartFromFile.Spec.DisplayName).ShouldNot(BeEmpty())
				Expect(quickStartFromFile.Spec.Description).ShouldNot(BeEmpty())
				Expect(quickStartFromFile.Spec.Icon).ShouldNot(BeEmpty())
				Expect(quickStartFromFile.Spec.DurationMinutes).ShouldNot(BeNil())
				Expect(quickStartFromFile.Spec.DurationMinutes).ShouldNot(BeZero())
				Expect(quickStartFromFile.Spec.Introduction).ShouldNot(BeEmpty())
				Expect(quickStartFromFile.Spec.Tasks).ShouldNot(BeEmpty())
				Expect(quickStartFromFile.Spec.Conclusion).ShouldNot(BeEmpty())
			}
		})
	})
})
