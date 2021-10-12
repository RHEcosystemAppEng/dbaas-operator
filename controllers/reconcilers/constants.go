package reconcilers

const (
	INSTALL_NAMESPACE                        = "openshift-operators"
	CATALOG_NAMESPACE                        = "openshift-marketplace"
	MONGODB_ATLAS_CATALOG_IMG                = "quay.io/ecosystem-appeng/mongodb-atlas-operator-catalog:0.7.1-dev"
	CRUNCHY_BRIDGE_CATALOG_IMG               = "quay.io/ecosystem-appeng/crunchy-bridge-operator-catalog:v0.0.2-dev"
	DBAAS_DYNAMIC_PLUGIN_IMG                 = "quay.io/ecosystem-appeng/dbaas-dynamic-plugin:0.1.3"
	DBAAS_DYNAMIC_PLUGIN_NAME                = "dbaas-dynamic-plugin"
	DBAAS_DYNAMIC_PLUGIN_DISPLAY_NAME        = "OpenShift Database as a Service Dynamic Plugin"
	CONSOLE_TELEMETRY_PLUGIN_IMG             = "quay.io/ecosystem-appeng/console-telemetry-plugin:0.1.3"
	CONSOLE_TELEMETRY_PLUGIN_NAME            = "console-telemetry-plugin"
	CONSOLE_TELEMETRY_PLUGIN_DISPLAY_NAME    = "Telemetry Plugin"
	CONSOLE_TELEMETRY_PLUGIN_SEGMENT_KEY_ENV = "SEGMENT_KEY"
	DBAAS_OPERATOR_VERSION_KEY_ENV           = "DBAAS_OPERATOR_VERSION"
	CONSOLE_TELEMETRY_PLUGIN_SEGMENT_KEY     = "qejcCDG37ICCLIDsM1FcJDkd68hglCoK"
	MONGODB_ATLAS_CSV                        = "mongodb-atlas-kubernetes.v0.7.1-dev"
	CRUNCHY_BRIDGE_CSV                       = "crunchy-bridge-operator.v0.0.2-dev"
	SERVICE_BINDING_CSV                      = "service-binding-operator.v0.10.0"
)
