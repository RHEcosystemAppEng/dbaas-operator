package reconcilers

import (
	corev1 "k8s.io/api/core/v1"

	dbaasv1alpha1 "github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
)

const (
	INSTALL_NAMESPACE              = "openshift-operators"
	CATALOG_NAMESPACE              = "openshift-marketplace"
	DBAAS_OPERATOR_VERSION_KEY_ENV = "DBAAS_OPERATOR_VERSION"
	CONSOLE_PLUGIN_49_TAG          = "-4.9"

	// CRUNCHY_BRIDGE
	CRUNCHY_BRIDGE_CATALOG_IMG = "registry.developers.crunchydata.com/crunchydata/crunchy-bridge-operator-catalog:v0.0.2"
	CRUNCHY_BRIDGE_CSV         = "crunchy-bridge-operator.v0.0.2"
	CRUNCHY_BRIDGE_NAME        = "crunchy-bridge"
	CRUNCHY_BRIDGE_DISPLAYNAME = "Crunchy Bridge Operator"
	CRUNCHY_BRIDGE_DEPLOYMENT  = "crunchy-bridge-operator-controller-manager"
	CRUNCHY_BRIDGE_PKG         = "crunchy-bridge-operator"
	CRUNCHY_BRIDGE_CHANNEL     = "alpha"

	// MONGODB_ATLAS
	MONGODB_ATLAS_CATALOG_IMG = "quay.io/mongodb/mongodb-atlas-kubernetes-dbaas-catalog:0.2.0"
	MONGODB_ATLAS_CSV         = "mongodb-atlas-kubernetes.v0.2.0"
	MONGODB_ATLAS_NAME        = "mongodb-atlas"
	MONGODB_ATLAS_DISPLAYNAME = "MongoDB Atlas Operator"
	MONGODB_ATLAS_DEPLOYMENT  = "mongodb-atlas-operator"
	MONGODB_ATLAS_PKG         = "mongodb-atlas-kubernetes"
	MONGODB_ATLAS_CHANNEL     = "beta"

	// COCKROACHDB
	COCKROACHDB_CSV         = "ccapi-k8s-operator.v0.0.1"
	COCKROACHDB_CATALOG_IMG = "gcr.io/cockroach-shared/ccapi-k8s-operator-catalog:v0.0.1"
	COCKROACHDB_NAME        = "ccapi-k8s"
	COCKROACHDB_DISPLAYNAME = "CockroachDB Cloud Operator"
	COCKROACHDB_DEPLOYMENT  = "ccapi-k8s-operator-controller-manager"
	COCKROACHDB_PKG         = "ccapi-k8s-operator"
	COCKROACHDB_CHANNEL     = "alpha"

	// DBAAS_DYNAMIC_PLUGIN
	DBAAS_DYNAMIC_PLUGIN_IMG          = "quay.io/ecosystem-appeng/dbaas-dynamic-plugin:0.1.5"
	DBAAS_DYNAMIC_PLUGIN_NAME         = "dbaas-dynamic-plugin"
	DBAAS_DYNAMIC_PLUGIN_DISPLAY_NAME = "OpenShift Database as a Service Dynamic Plugin"

	// CONSOLE_TELEMETRY_PLUGIN
	CONSOLE_TELEMETRY_PLUGIN_IMG             = "quay.io/ecosystem-appeng/console-telemetry-plugin:0.1.4"
	CONSOLE_TELEMETRY_PLUGIN_NAME            = "console-telemetry-plugin"
	CONSOLE_TELEMETRY_PLUGIN_DISPLAY_NAME    = "Telemetry Plugin"
	CONSOLE_TELEMETRY_PLUGIN_SEGMENT_KEY_ENV = "SEGMENT_KEY"
	CONSOLE_TELEMETRY_PLUGIN_SEGMENT_KEY     = "qejcCDG37ICCLIDsM1FcJDkd68hglCoK"
)

var InstallationPlatforms = map[dbaasv1alpha1.PlatformsName]dbaasv1alpha1.PlatformConfig{
	dbaasv1alpha1.DBaaSDynamicPluginInstallation: {
		Name:        DBAAS_DYNAMIC_PLUGIN_NAME,
		Image:       DBAAS_DYNAMIC_PLUGIN_IMG,
		DisplayName: DBAAS_DYNAMIC_PLUGIN_DISPLAY_NAME,
		Type:        dbaasv1alpha1.TypeConsolePlugin,
	},
	dbaasv1alpha1.ConsoleTelemetryPluginInstallation: {
		Name:        CONSOLE_TELEMETRY_PLUGIN_NAME,
		Image:       CONSOLE_TELEMETRY_PLUGIN_IMG,
		DisplayName: CONSOLE_TELEMETRY_PLUGIN_DISPLAY_NAME,
		Envs:        []corev1.EnvVar{{Name: CONSOLE_TELEMETRY_PLUGIN_SEGMENT_KEY_ENV, Value: CONSOLE_TELEMETRY_PLUGIN_SEGMENT_KEY}},
		Type:        dbaasv1alpha1.TypeConsolePlugin,
	},
	dbaasv1alpha1.CrunchyBridgeInstallation: {
		Name:           CRUNCHY_BRIDGE_NAME,
		CSV:            CRUNCHY_BRIDGE_CSV,
		DeploymentName: CRUNCHY_BRIDGE_DEPLOYMENT,
		Image:          CRUNCHY_BRIDGE_CATALOG_IMG,
		PackageName:    CRUNCHY_BRIDGE_PKG,
		Channel:        CRUNCHY_BRIDGE_CHANNEL,
		DisplayName:    CRUNCHY_BRIDGE_DISPLAYNAME,
		Type:           dbaasv1alpha1.TypeProvider,
	},
	dbaasv1alpha1.MongoDBAtlasInstallation: {
		Name:           MONGODB_ATLAS_NAME,
		CSV:            MONGODB_ATLAS_CSV,
		DeploymentName: MONGODB_ATLAS_DEPLOYMENT,
		Image:          MONGODB_ATLAS_CATALOG_IMG,
		PackageName:    MONGODB_ATLAS_PKG,
		Channel:        MONGODB_ATLAS_CHANNEL,
		DisplayName:    MONGODB_ATLAS_DISPLAYNAME,
		Type:           dbaasv1alpha1.TypeProvider,
	},
	dbaasv1alpha1.CockroachDBInstallation: {
		Name:           COCKROACHDB_NAME,
		CSV:            COCKROACHDB_CSV,
		DeploymentName: COCKROACHDB_DEPLOYMENT,
		Image:          COCKROACHDB_CATALOG_IMG,
		PackageName:    COCKROACHDB_PKG,
		Channel:        COCKROACHDB_CHANNEL,
		DisplayName:    COCKROACHDB_DISPLAYNAME,
		Type:           dbaasv1alpha1.TypeProvider,
	},
	dbaasv1alpha1.DBaaSQuickStartInstallation: {
		Type: dbaasv1alpha1.TypeQuickStart,
	},
}
