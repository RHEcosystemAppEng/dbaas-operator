module github.com/RHEcosystemAppEng/dbaas-operator

go 1.16

require (
	github.com/go-logr/logr v1.2.0
	github.com/imdario/mergo v0.3.12
	github.com/oklog/run v1.1.0
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.13.0
	github.com/openshift/api v0.0.0-20210910062324-a41d3573a3ba
	github.com/openshift/client-go v0.0.0-20200320143156-e7fa42a1261e
	github.com/operator-framework/api v0.10.5
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.54.1
	github.com/prometheus/client_golang v1.11.0
	github.com/spf13/pflag v1.0.5
	go.uber.org/zap v1.19.0
	k8s.io/api v0.23.0
	k8s.io/apimachinery v0.23.0
	k8s.io/client-go v0.23.0
	k8s.io/klog v1.0.0
	k8s.io/utils v0.0.0-20210930125809-cb0fa318a74b
	sigs.k8s.io/controller-runtime v0.9.0
)
