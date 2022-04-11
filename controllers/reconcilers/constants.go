package reconcilers

import (
	"os"

	dbaasv1alpha1 "github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
	embeddedconfigs "github.com/RHEcosystemAppEng/dbaas-operator/config"
	"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	INSTALL_NAMESPACE              = "openshift-operators"
	CATALOG_NAMESPACE              = "openshift-marketplace"
	DBAAS_OPERATOR_VERSION_KEY_ENV = "DBAAS_OPERATOR_VERSION"
	CONSOLE_PLUGIN_49_TAG          = "-4.9"

	// CRUNCHY_BRIDGE
	CRUNCHY_BRIDGE_CATALOG_IMG = "RELATED_IMAGE_CRUNCHY_BRIDGE_CATALOG"
	CRUNCHY_BRIDGE_CSV         = "CSV_VERSION_CRUNCHY_BRIDGE"
	CRUNCHY_BRIDGE_NAME        = "crunchy-bridge"
	CRUNCHY_BRIDGE_DISPLAYNAME = "Crunchy Bridge Operator"
	CRUNCHY_BRIDGE_DEPLOYMENT  = "crunchy-bridge-operator-controller-manager"
	CRUNCHY_BRIDGE_PKG         = "crunchy-bridge-operator"
	CRUNCHY_BRIDGE_CHANNEL     = "alpha"

	// MONGODB_ATLAS
	MONGODB_ATLAS_CATALOG_IMG = "RELATED_IMAGE_MONGODB_ATLAS_CATALOG"
	MONGODB_ATLAS_CSV         = "CSV_VERSION_MONGODB_ATLAS"
	MONGODB_ATLAS_NAME        = "mongodb-atlas"
	MONGODB_ATLAS_DISPLAYNAME = "MongoDB Atlas Operator"
	MONGODB_ATLAS_DEPLOYMENT  = "mongodb-atlas-operator"
	MONGODB_ATLAS_PKG         = "mongodb-atlas-kubernetes"
	MONGODB_ATLAS_CHANNEL     = "beta"

	// COCKROACHDB
	COCKROACHDB_CSV         = "CSV_VERSION_COCKROACHDB"
	COCKROACHDB_CATALOG_IMG = "RELATED_IMAGE_COCKROACHDB_CATALOG"
	COCKROACHDB_NAME        = "ccapi-k8s"
	COCKROACHDB_DISPLAYNAME = "CockroachDB Cloud Operator"
	COCKROACHDB_DEPLOYMENT  = "ccapi-k8s-operator-controller-manager"
	COCKROACHDB_PKG         = "ccapi-k8s-operator"
	COCKROACHDB_CHANNEL     = "alpha"

	// DBAAS_DYNAMIC_PLUGIN
	DBAAS_DYNAMIC_PLUGIN_IMG          = "RELATED_IMAGE_DBAAS_DYNAMIC_PLUGIN"
	DBAAS_DYNAMIC_PLUGIN_VERSION      = "CSV_VERSION_DBAAS_DYNAMIC_PLUGIN"
	DBAAS_DYNAMIC_PLUGIN_NAME         = "dbaas-dynamic-plugin"
	DBAAS_DYNAMIC_PLUGIN_DISPLAY_NAME = "OpenShift Database as a Service Dynamic Plugin"

	// CONSOLE_TELEMETRY_PLUGIN
	CONSOLE_TELEMETRY_PLUGIN_IMG             = "RELATED_IMAGE_CONSOLE_TELEMETRY_PLUGIN"
	CONSOLE_TELEMETRY_PLUGIN_VERSION         = "CSV_VERSION_CONSOLE_TELEMETRY_PLUGIN"
	CONSOLE_TELEMETRY_PLUGIN_NAME            = "console-telemetry-plugin"
	CONSOLE_TELEMETRY_PLUGIN_DISPLAY_NAME    = "Telemetry Plugin"
	CONSOLE_TELEMETRY_PLUGIN_SEGMENT_KEY_ENV = "SEGMENT_KEY"
	CONSOLE_TELEMETRY_PLUGIN_SEGMENT_KEY     = "qejcCDG37ICCLIDsM1FcJDkd68hglCoK"

	// RDS_PROVIDER
	RDS_PROVIDER_CSV         = "CSV_VERSION_RDS_PROVIDER"
	RDS_PROVIDER_CATALOG_IMG = "RELATED_IMAGE_RDS_PROVIDER_CATALOG"
	RDS_PROVIDER_NAME        = "rds-provider"
	RDS_PROVIDER_DISPLAYNAME = "RHODA Provider Operator for Amazon RDS"
	RDS_PROVIDER_DEPLOYMENT  = "rds-dbaas-operator-controller-manager"
	RDS_PROVIDER_PKG         = "rds-dbaas-operator"
	RDS_PROVIDER_CHANNEL     = "alpha"

	// OBSERVABILITY
	OBSERVABILITY_CATALOG_IMG = "RELATED_IMAGE_OBSERVABILITY_CATALOG"
	OBSERVABILITY_CSV         = "CSV_VERSION_OBSERVABILITY"
	OBSERVABILITY_NAME        = "observability"
	OBSERVABILITY_DISPLAYNAME = "observability Operator"
	OBSERVABILITY_DEPLOYMENT  = "observability-operator"
	OBSERVABILITY_PKG         = "observability-operator"
	OBSERVABILITY_CHANNEL     = "stable"
)

var InstallationPlatforms = map[dbaasv1alpha1.PlatformsName]dbaasv1alpha1.PlatformConfig{
	dbaasv1alpha1.DBaaSDynamicPluginInstallation: {
		Name:        DBAAS_DYNAMIC_PLUGIN_NAME,
		Image:       fetchEnvValue(DBAAS_DYNAMIC_PLUGIN_IMG),
		DisplayName: DBAAS_DYNAMIC_PLUGIN_DISPLAY_NAME,
		CSV:         fetchEnvValue(DBAAS_DYNAMIC_PLUGIN_VERSION),
		Type:        dbaasv1alpha1.TypeConsolePlugin,
	},
	dbaasv1alpha1.ConsoleTelemetryPluginInstallation: {
		Name:        CONSOLE_TELEMETRY_PLUGIN_NAME,
		Image:       fetchEnvValue(CONSOLE_TELEMETRY_PLUGIN_IMG),
		DisplayName: CONSOLE_TELEMETRY_PLUGIN_DISPLAY_NAME,
		CSV:         fetchEnvValue(CONSOLE_TELEMETRY_PLUGIN_VERSION),
		Envs:        []corev1.EnvVar{{Name: CONSOLE_TELEMETRY_PLUGIN_SEGMENT_KEY_ENV, Value: CONSOLE_TELEMETRY_PLUGIN_SEGMENT_KEY}},
		Type:        dbaasv1alpha1.TypeConsolePlugin,
	},
	dbaasv1alpha1.CrunchyBridgeInstallation: {
		Name:           CRUNCHY_BRIDGE_NAME,
		CSV:            fetchEnvValue(CRUNCHY_BRIDGE_CSV),
		DeploymentName: CRUNCHY_BRIDGE_DEPLOYMENT,
		Image:          fetchEnvValue(CRUNCHY_BRIDGE_CATALOG_IMG),
		PackageName:    CRUNCHY_BRIDGE_PKG,
		Channel:        CRUNCHY_BRIDGE_CHANNEL,
		DisplayName:    CRUNCHY_BRIDGE_DISPLAYNAME,
		Type:           dbaasv1alpha1.TypeOperator,
	},
	dbaasv1alpha1.MongoDBAtlasInstallation: {
		Name:           MONGODB_ATLAS_NAME,
		CSV:            fetchEnvValue(MONGODB_ATLAS_CSV),
		DeploymentName: MONGODB_ATLAS_DEPLOYMENT,
		Image:          fetchEnvValue(MONGODB_ATLAS_CATALOG_IMG),
		PackageName:    MONGODB_ATLAS_PKG,
		Channel:        MONGODB_ATLAS_CHANNEL,
		DisplayName:    MONGODB_ATLAS_DISPLAYNAME,
		Type:           dbaasv1alpha1.TypeOperator,
	},
	dbaasv1alpha1.CockroachDBInstallation: {
		Name:           COCKROACHDB_NAME,
		CSV:            fetchEnvValue(COCKROACHDB_CSV),
		DeploymentName: COCKROACHDB_DEPLOYMENT,
		Image:          fetchEnvValue(COCKROACHDB_CATALOG_IMG),
		PackageName:    COCKROACHDB_PKG,
		Channel:        COCKROACHDB_CHANNEL,
		DisplayName:    COCKROACHDB_DISPLAYNAME,
		Type:           dbaasv1alpha1.TypeOperator,
	},
	dbaasv1alpha1.DBaaSQuickStartInstallation: {
		Type: dbaasv1alpha1.TypeQuickStart,
	},
	dbaasv1alpha1.RDSProviderInstallation: {
		Name:           RDS_PROVIDER_NAME,
		CSV:            fetchEnvValue(RDS_PROVIDER_CSV),
		DeploymentName: RDS_PROVIDER_DEPLOYMENT,
		Image:          fetchEnvValue(RDS_PROVIDER_CATALOG_IMG),
		PackageName:    RDS_PROVIDER_PKG,
		Channel:        RDS_PROVIDER_CHANNEL,
		DisplayName:    RDS_PROVIDER_DISPLAYNAME,
		Type:           dbaasv1alpha1.TypeOperator,
	},
	dbaasv1alpha1.ObservabilityInstallation: {
		Name:           OBSERVABILITY_NAME,
		CSV:            fetchEnvValue(OBSERVABILITY_CSV),
		DeploymentName: OBSERVABILITY_DEPLOYMENT,
		Image:          fetchEnvValue(OBSERVABILITY_CATALOG_IMG),
		PackageName:    OBSERVABILITY_PKG,
		Channel:        OBSERVABILITY_CHANNEL,
		DisplayName:    OBSERVABILITY_DISPLAYNAME,
		Type:           dbaasv1alpha1.TypeOperator,
	},
}

// fetchEnvValue returns the value of a set variable. if env var not set, returns the
// 		default value from an embedded yaml file.
func fetchEnvValue(envVar string) (imageValue string) {
	imageValue, found := os.LookupEnv(envVar)
	if !found {
		tmpDep := &appsv1.Deployment{}
		if err := yaml.Unmarshal(embeddedconfigs.EnvImages, tmpDep); err != nil {
			return imageValue
		}
		return getEnvVarValue(envVar, tmpDep.Spec.Template.Spec.Containers[0].Env)
	}
	return imageValue
}

// getEnvVarValue returns the value of an EnvVar by name
func getEnvVarValue(envName string, env []corev1.EnvVar) string {
	for _, v := range env {
		if v.Name == envName {
			return v.Value
		}
	}
	return ""
}
