# VERSION defines the project version for the bundle.
# Update this value when you upgrade the version of your project.
# To re-generate a bundle for another specific version without changing the standard setup, you can:
# - use the VERSION as arg of the bundle target (e.g make bundle VERSION=0.0.2)
# - use environment variables to overwrite this value (e.g export VERSION=0.0.2)
VERSION ?= 0.5.0

CONTAINER_ENGINE?=docker

DEV ?= false
ifeq ($(DEV),true)
  VERSION := $(VERSION)-$(shell git rev-parse --short HEAD)
endif

LOCALDIR ?= $(shell pwd)
## Location to install dependencies to
LOCALBIN ?= $(LOCALDIR)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
CRD_REF_DOCS ?= $(LOCALBIN)/crd-ref-docs
ENVTEST ?= $(LOCALBIN)/setup-envtest

## Tool Versions
OPM_VERSION ?= v1.26.2
OPERATOR_SDK_VERSION ?= v1.22.2
CONTROLLER_TOOLS_VERSION ?= v0.4.1
ENVTEST_K8S_VERSION ?= 1.25.0

# OLD_BUNDLE_VERSIONS defines the comma separated list of versions of old bundles to add to the index.
#
# This is NOT required if you are incrementally building the catalog index. If you need to increment on
# a catalog that already HAS old bundles, such as when building/pushing the official release, then you do NOT
# need to uncomment & add the old bundles - the existing "from-index" catalog already has those.
#
# If you are developing and pushing against your OWN quay for testing, you likely need to uncomment
#OLD_BUNDLE_VERSIONS ?= 0.1.0,0.1.1,0.1.2,0.1.3

# ORG indicates the organization that docker images will be build for & pushed to
# CHANGE THIS TO YOUR OWN QUAY USERNAME FOR DEV/TESTING/PUSHING
ORG ?= ecosystem-appeng

# CATALOG_BASE_IMG defines an existing catalog version to build on & add bundles to
# CATALOG_BASE_IMG ?= quay.io/$(ORG)/dbaas-operator-catalog:v$(VERSION)
CATALOG_BASE_IMG ?= quay.io/ecosystem-appeng/dbaas-operator-catalog:0.4.0-wrapper

export OPERATOR_CONDITION_NAME=dbaas-operator.v$(VERSION)

# CHANNELS define the bundle channels used in the bundle.
# Add a new line here if you would like to change its default config. (E.g CHANNELS = "preview,fast,stable")
# To re-generate a bundle for other specific channels without changing the standard setup, you can:
# - use the CHANNELS as arg of the bundle target (e.g make bundle CHANNELS=preview,fast,stable)
# - use environment variables to overwrite this value (e.g export CHANNELS="preview,fast,stable")
ifneq ($(origin CHANNELS), undefined)
BUNDLE_CHANNELS := --channels=$(CHANNELS)
endif

# DEFAULT_CHANNEL defines the default channel used in the bundle.
# Add a new line here if you would like to change its default config. (E.g DEFAULT_CHANNEL = "stable")
# To re-generate a bundle for any other default channel without changing the default setup, you can:
# - use the DEFAULT_CHANNEL as arg of the bundle target (e.g make bundle DEFAULT_CHANNEL=stable)
# - use environment variables to overwrite this value (e.g export DEFAULT_CHANNEL="stable")
ifneq ($(origin DEFAULT_CHANNEL), undefined)
BUNDLE_DEFAULT_CHANNEL := --default-channel=$(DEFAULT_CHANNEL)
endif
BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

# IMAGE_TAG_BASE defines the docker.io namespace and part of the image name for remote images.
# This variable is used to construct full image tags for bundle and catalog images.
#
# For example, running 'make bundle-build bundle-push catalog-build catalog-push' will build and push both
# redhat.com/dbaas-operator-bundle:$VERSION and redhat.com/dbaas-operator-catalog:$VERSION.
IMAGE_TAG_BASE ?= quay.io/$(ORG)/dbaas-operator

# BUNDLE_IMG defines the image:tag used for the bundle.
# You can use it as an arg. (E.g make bundle-build BUNDLE_IMG=<some-registry>/<project-name-bundle>:<tag>)
BUNDLE_IMG ?= $(IMAGE_TAG_BASE)-bundle:v$(VERSION)

# OLD_BUNDLE_IMGS defines the comma separated list of old bundles to add to the index.
COMMA := ,
EMPTY :=
SPACE := $(EMPTY) $(EMPTY)
OLD_BUNDLE_IMG_TAG_BASE ?= $(IMAGE_TAG_BASE)-bundle
OLD_BUNDLE_IMGS ?= $(patsubst %$(COMMA),%$(EMPTY),$(subst $(SPACE),$(EMPTY),$(foreach ver,$(subst $(COMMA),$(SPACE),$(OLD_BUNDLE_VERSIONS)),$(OLD_BUNDLE_IMG_TAG_BASE):v$(ver),)))

# Image URL to use all building/pushing image targets
IMG ?= $(IMAGE_TAG_BASE):v$(VERSION)
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true,preserveUnknownFields=false"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

all: build


##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

install-tools:
	go install golang.org/x/tools/cmd/goimports@v0.1.11
	go install github.com/mgechev/revive@v1.2.1

fmt: install-tools
	goimports -w .

lint: install-tools
	goimports -d .
	revive -config ./config.toml ./...

vet: ## Run go vet against code.
	go vet ./...

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	test -s $(LOCALBIN)/setup-envtest || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

.PHONY: test
test: sdk-manifests vet envtest ## Run tests.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" go test ./... -coverprofile cover.tmp.out -covermode count

.PHONY: coverage
coverage: test
	cat cover.tmp.out | grep -v "_generated.*.go" > cover.out
	go tool cover -func=cover.out

##@ Build
release-build: bundle docker-build bundle-build bundle-push catalog-build ## Build operator docker, bundle, catalog images

build: generate fmt vet ## Build manager binary.
	go build -o bin/manager main.go

run: sdk-manifests vet ## Run a controller from your host.
	go run ./main.go

docker-build: test ## Build docker image with the manager.
	$(CONTAINER_ENGINE) build --pull --platform linux/amd64 -t ${IMG} .

docker-push: ## Push docker image with the manager.
	$(CONTAINER_ENGINE) push ${IMG}

##@ Deployment

release-push: docker-push bundle-push catalog-push ## Push operator docker, bundle, catalog images

install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl delete -f -

deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/default | kubectl delete -f -

deploy-olm:
	oc apply -f config/samples/catalog-operator-group.yaml
	oc apply -f config/samples/catalog-subscription.yaml

undeploy-olm:
	-oc delete subscriptions.operators.coreos.com dbaas-operator
	-oc delete operatorgroup dbaas-operator-group
	-oc delete clusterserviceversion dbaas-operator.v${VERSION}

catalog-update:
	-oc delete catalogsource dbaas-operator -n openshift-marketplace
	-oc delete catalogsource crunchy-bridge-catalogsource -n openshift-marketplace
	-oc delete catalogsource ccapi-k8s-catalogsource -n openshift-marketplace
	-oc delete catalogsource observability-catalogsource -n openshift-marketplace
	-oc delete catalogsource rds-provider-catalogsource -n openshift-marketplace
	 oc apply -f config/samples/catalog-source.yaml

deploy-sample-app:
	oc apply -f config/samples/quarkus-runner/deployment.yaml

undeploy-sample-app:
	-oc delete servicebindings.binding.operators.coreos.com dbaas-quarkus-sample-app-d-atlas-connection-dbsc
	-oc delete deployment dbaas-quarkus-sample-app
	-oc delete service dbaas-quarkus-sample-app
	-oc delete route dbaas-quarkus-sample-app

deploy-sample-binding:
	oc apply -f config/samples/quarkus-runner/sample-binding.yaml

undeploy-sample-binding:
	oc delete servicebindings.binding.operators.coreos.com dbaas-quarkus-sample-app-d-atlas-connection-dbsc

clean-namespace:
	-oc delete dbaasconnections.dbaas.redhat.com --all
	-oc delete dbaasservices.dbaas.redhat.com --all

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: crd-ref-docs
crd-ref-docs: $(CRD_REF_DOCS) ## Download crd-ref-docs locally if necessary.
$(CRD_REF_DOCS): $(LOCALBIN)
	test -s $(LOCALBIN)/crd-ref-docs || GOBIN=$(LOCALBIN) go install github.com/elastic/crd-ref-docs@v0.0.8

.PHONY: generate-ref
generate-ref: generate fmt crd-ref-docs
	$(CRD_REF_DOCS) --log-level=WARN --config=$(LOCALDIR)/ref-templates/config.yaml --source-path=$(LOCALDIR)/api/v1beta1 --renderer=asciidoctor --templates-dir=$(LOCALDIR)/ref-templates/asciidoctor --output-path=$(LOCALDIR)/docs/api/asciidoc/ref.adoc
	$(CRD_REF_DOCS) --log-level=WARN --config=$(LOCALDIR)/ref-templates/config.yaml --source-path=$(LOCALDIR)/api/v1beta1 --renderer=markdown --templates-dir=$(LOCALDIR)/ref-templates/markdown --output-path=$(LOCALDIR)/docs/api/markdown/ref.md

.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	test -s $(LOCALBIN)/kustomize || GOBIN=$(LOCALBIN) go install sigs.k8s.io/kustomize/kustomize/v4@v4.5.7

# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go get $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef

.PHONY: sdk-manifests
sdk-manifests: manifests generate-ref kustomize sdk ## Generate bundle manifests and metadata.
	$(SDK) generate kustomize manifests -q
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(IMG)

.PHONY: bundle
bundle: sdk-manifests ## Generate bundle manifests, then validate generated files.
	$(KUSTOMIZE) build config/manifests | $(SDK) generate bundle -q --overwrite --manifests --version $(VERSION) $(BUNDLE_METADATA_OPTS)
	$(SDK) bundle validate ./bundle --select-optional suite=operatorframework

.PHONY: bundle-w-digests
bundle-w-digests: sdk-manifests ## Generate bundle manifests w/ image digests, then validate generated files.
	$(KUSTOMIZE) build config/manifests | $(SDK) generate bundle -q --overwrite --manifests --use-image-digests --version $(VERSION) $(BUNDLE_METADATA_OPTS)
	$(SDK) bundle validate ./bundle --select-optional suite=operatorframework

.PHONY: bundle-build
bundle-build: ## Build the bundle image.
	$(CONTAINER_ENGINE) build -f bundle.Dockerfile --platform linux/amd64 -t $(BUNDLE_IMG) .

.PHONY: bundle-push
bundle-push: ## Push the bundle image.
	$(MAKE) docker-push IMG=$(BUNDLE_IMG)

.PHONY: opm
OPM = ./bin/opm.$(OPM_VERSION)
opm: ## Download opm locally if necessary.
ifeq (,$(wildcard $(OPM)))
	@{ \
	set -e ;\
	mkdir -p $(dir $(OPM)) ;\
	OS=$(shell go env GOOS) && ARCH=$(shell go env GOARCH) && \
	curl -sSLo $(OPM) https://github.com/operator-framework/operator-registry/releases/download/$(OPM_VERSION)/$${OS}-$${ARCH}-opm ;\
	chmod +x $(OPM) ;\
	}
endif

.PHONY: sdk
SDK = ./bin/operator-sdk.$(OPERATOR_SDK_VERSION)
sdk: ## Download operator-sdk if necessary.
ifeq (,$(wildcard $(SDK)))
	@{ \
	set -e ;\
	mkdir -p $(dir $(SDK)) ;\
	OS=$(shell go env GOOS) && ARCH=$(shell go env GOARCH) && \
	curl -sSLo $(SDK) https://github.com/operator-framework/operator-sdk/releases/download/$(OPERATOR_SDK_VERSION)/operator-sdk_$${OS}_$${ARCH} ;\
	chmod +x $(SDK) ;\
	}
endif

# A comma-separated list of bundle images (e.g. make catalog-build BUNDLE_IMGS=example.com/operator-bundle:v0.1.0,example.com/operator-bundle:v0.2.0).
# These images MUST exist in a registry and be pull-able.
ifeq ($(OLD_BUNDLE_IMGS),)
BUNDLE_IMGS ?= $(BUNDLE_IMG)
else
BUNDLE_IMGS ?= $(BUNDLE_IMG),$(OLD_BUNDLE_IMGS)
endif

# The image tag given to the resulting catalog image (e.g. make catalog-build CATALOG_IMG=example.com/operator-catalog:v0.2.0).
CATALOG_IMG ?= $(IMAGE_TAG_BASE)-catalog:v$(VERSION)

# Set CATALOG_BASE_IMG to an existing catalog image tag to add $BUNDLE_IMGS to that image.
ifneq ($(origin CATALOG_BASE_IMG), undefined)
FROM_INDEX_OPT := --from-index $(CATALOG_BASE_IMG)
endif

# Build a catalog image by adding bundle images to an empty catalog using the operator package manager tool, 'opm'.
# This recipe invokes 'opm' in 'semver' bundle add mode. For more information on add modes, see:
# https://github.com/operator-framework/community-operators/blob/7f1438c/docs/packaging-operator.md#updating-your-existing-operator
.PHONY: catalog-build
catalog-build: opm ## Build a catalog image.
	$(OPM) index add --container-tool $(CONTAINER_ENGINE) --mode semver --tag $(CATALOG_IMG) --bundles $(BUNDLE_IMGS)
#	$(OPM) index add --container-tool $(CONTAINER_ENGINE) --mode semver --tag $(CATALOG_IMG) --bundles $(BUNDLE_IMGS) $(FROM_INDEX_OPT)

# Push the catalog image.
.PHONY: catalog-push
catalog-push: ## Push a catalog image.
	$(MAKE) docker-push IMG=$(CATALOG_IMG)

.PHONY: wrapper-build
wrapper-build: ## Build the catalog wrapper image.
	$(CONTAINER_ENGINE) build --pull -f wrapper.Dockerfile --platform linux/amd64 -t quay.io/$(ORG)/dbaas-operator-catalog:0.4.0-wrapper .

.PHONY: wrapper-push
wrapper-push: ## Push the catalog wrapper image.
	$(MAKE) docker-push IMG=quay.io/$(ORG)/dbaas-operator-catalog:0.4.0-wrapper

.PHONY: get-version
get-version: ; $(info ${VERSION})
	@echo -n
