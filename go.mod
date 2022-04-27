module github.com/RHEcosystemAppEng/dbaas-operator

go 1.16

require (
	github.com/RHsyseng/operator-utils v1.4.9
	github.com/go-logr/logr v0.4.0
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.15.0
	github.com/openshift/api v0.0.0-20210910062324-a41d3573a3ba
	github.com/operator-framework/api v0.10.5
	github.com/prometheus/client_golang v1.11.0
	go.uber.org/zap v1.19.0
	golang.org/x/mod v0.5.1
	k8s.io/api v0.22.1
	k8s.io/apimachinery v0.22.1
	k8s.io/client-go v0.22.1
	k8s.io/utils v0.0.0-20210802155522-efc7438f0176
	sigs.k8s.io/controller-runtime v0.10.0
)
