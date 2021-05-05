module github.com/RHEcosystemAppEng/dbaas-operator

go 1.15

require (
	github.com/go-logr/logr v0.4.0
	github.com/mongodb/mongodb-atlas-kubernetes v0.5.0
	github.com/onsi/ginkgo v1.15.2
	github.com/onsi/gomega v1.11.0
	k8s.io/api v0.19.2
	k8s.io/apimachinery v0.19.2
	k8s.io/client-go v0.19.2
	k8s.io/utils v0.0.0-20200912215256-4140de9c8800
	sigs.k8s.io/controller-runtime v0.7.2
)

replace github.com/mongodb/mongodb-atlas-kubernetes => github.com/jeremyary/mongodb-atlas-kubernetes v1.0.4
