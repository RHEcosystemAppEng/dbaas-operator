package reconcilers

const (
	INSTALL_NAMESPACE              = "openshift-operators"
	CATALOG_NAMESPACE              = "openshift-marketplace"
	DBAAS_OPERATOR_VERSION_KEY_ENV = "DBAAS_OPERATOR_VERSION"

	//CRUNCHY_BRIDGE
	CRUNCHY_BRIDGE_CATALOG_IMG = "quay.io/ecosystem-appeng/crunchy-bridge-operator-catalog:v0.0.2-dev"
	CRUNCHY_BRIDGE_CSV         = "crunchy-bridge-operator.v0.0.2-dev"
	CRUNCHY_BRIDGE_NAME        = "crunchy-bridge"
	CRUNCHY_BRIDGE_DISPLAYNAME = "Crunchy Bridge Operator"
	CRUNCHY_BRIDGE_DEPLOYMENT  = "crunchy-bridge-operator-controller-manager"
	CRUNCHY_BRIDGE_PKG         = "crunchy-bridge-operator"
	CRUNCHY_BRIDGE_CHANNEL     = "alpha"

	//MONGODB_ATLAS
	MONGODB_ATLAS_CATALOG_IMG = "quay.io/ecosystem-appeng/mongodb-atlas-operator-catalog:0.7.1-dev"
	MONGODB_ATLAS_CSV         = "mongodb-atlas-kubernetes.v0.7.1-dev"
	MONGODB_ATLAS_NAME        = "mongodb-atlas"
	MONGODB_ATLAS_DISPLAYNAME = "MongoDB Atlas Operator"
	MONGODB_ATLAS_DEPLOYMENT  = "mongodb-atlas-operator"
	MONGODB_ATLAS_PKG         = "mongodb-atlas-kubernetes"
	MONGODB_ATLAS_CHANNEL     = "beta"

	//COCKROACHDB
	COCKROACHDB_CSV         = "ccapi-k8s-operator.v0.0.1"
	COCKROACHDB_CATALOG_IMG = "quay.io/ecosystem-appeng/ccapi-k8s-operator-catalog:v0.0.1"
	COCKROACHDB_NAME        = "ccapi-k8s"
	COCKROACHDB_DISPLAYNAME = "CockroachDB Cloud Operator"
	COCKROACHDB_DEPLOYMENT  = "ccapi-k8s-operator-controller-manager"
	COCKROACHDB_PKG         = "ccapi-k8s-operator"
	COCKROACHDB_CHANNEL     = "alpha"

	//DBAAS_DYNAMIC_PLUGIN
	DBAAS_DYNAMIC_PLUGIN_IMG          = "quay.io/ecosystem-appeng/dbaas-dynamic-plugin:0.1.4"
	DBAAS_DYNAMIC_PLUGIN_NAME         = "dbaas-dynamic-plugin"
	DBAAS_DYNAMIC_PLUGIN_DISPLAY_NAME = "OpenShift Database as a Service Dynamic Plugin"

	//CONSOLE_TELEMETRY_PLUGIN
	CONSOLE_TELEMETRY_PLUGIN_IMG             = "quay.io/ecosystem-appeng/console-telemetry-plugin:0.1.4"
	CONSOLE_TELEMETRY_PLUGIN_NAME            = "console-telemetry-plugin"
	CONSOLE_TELEMETRY_PLUGIN_DISPLAY_NAME    = "Telemetry Plugin"
	CONSOLE_TELEMETRY_PLUGIN_SEGMENT_KEY_ENV = "SEGMENT_KEY"
	CONSOLE_TELEMETRY_PLUGIN_SEGMENT_KEY     = "qejcCDG37ICCLIDsM1FcJDkd68hglCoK"
)
