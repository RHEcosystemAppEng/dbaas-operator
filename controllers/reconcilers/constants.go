package reconcilers

import (
	dbaasv1alpha1 "github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

const (
	// InstallNamespace namespace for installation
	InstallNamespace = "openshift-operators"
	// CatalogNamespace namespace for catalog sources
	CatalogNamespace = "openshift-marketplace"
	// ConsolePlugin49Tag tag for the console plugin in OpenShift 4.9
	ConsolePlugin49Tag = "-4.9"

	// CRUNCHY_BRIDGE
	crunchyBridgeCatlaogImg  = "registry.developers.crunchydata.com/crunchydata/crunchy-bridge-operator-catalog:v0.0.5"
	crunchyBridgeCSV         = "crunchy-bridge-operator.v0.0.5"
	crunchyBridgeName        = "crunchy-bridge"
	crunchyBridgeDisplayName = "Crunchy Bridge Operator"
	crunchyBridgeDeployment  = "crunchy-bridge-operator-controller-manager"
	crunchyBridgePkg         = "crunchy-bridge-operator"
	crunchyBridgeChannel     = "alpha"

	// MONGODB_ATLAS
	mongoDBAtlasCatalogImg  = "quay.io/mongodb/mongodb-atlas-kubernetes-dbaas-catalog:0.2.0"
	mongoDBAtlasCSV         = "mongodb-atlas-kubernetes.v0.2.0"
	mongoDBAtlasName        = "mongodb-atlas"
	mongoDBAtlasDisplayName = "MongoDB Atlas Operator"
	mongoDBAtlasDeployment  = "mongodb-atlas-operator"
	mongoDBAtlasPkg         = "mongodb-atlas-kubernetes"
	mongoDBAtlasChannel     = "beta"

	// COCKROACHDB
	cockroachDBCSV         = "ccapi-k8s-operator.v0.0.3"
	cockroachDBCatalogImg  = "gcr.io/cockroach-shared/ccapi-k8s-operator-catalog:v0.0.3"
	cockroachDBName        = "ccapi-k8s"
	cockroachDBDisplayName = "CockroachDB Cloud Operator"
	cockroachDBDeployment  = "ccapi-k8s-operator-controller-manager"
	cockroachDBPkg         = "ccapi-k8s-operator"
	cockroachDBChannel     = "alpha"

	// DBAAS_DYNAMIC_PLUGIN
	dbaassDynamicPluginImg         = "quay.io/ecosystem-appeng/dbaas-dynamic-plugin:0.2.0"
	dbaassDynamicPluginName        = "dbaas-dynamic-plugin"
	dbaassDynamicPluginDisplayName = "OpenShift Database as a Service Dynamic Plugin"

	// CONSOLE_TELEMETRY_PLUGIN
	consoleTelemetryPluginImg           = "quay.io/ecosystem-appeng/console-telemetry-plugin:0.1.4"
	consoleTelemetryPluginName          = "console-telemetry-plugin"
	consoleTelemetryPluginDisplayName   = "Telemetry Plugin"
	consoleTelemetryPluginSegmentKeyEnv = "SEGMENT_KEY"
	consoleTelemetryPluginSegmentKey    = "qejcCDG37ICCLIDsM1FcJDkd68hglCoK"

	// RDS_PROVIDER
	rdsProviderCSV         = "rds-dbaas-operator.v0.1.0"
	rdsProviderCatalogImg  = "quay.io/ecosystem-appeng/rds-dbaas-operator-catalog:v0.1.0"
	rdsProviderName        = "rds-provider"
	rdsProviderDisplayName = "RHODA Provider Operator for Amazon RDS"
	rdsProviderDeployment  = "rds-dbaas-operator-controller-manager"
	rdsProviderPkg         = "rds-dbaas-operator"
	rdsProviderChannel     = "alpha"
)

// InstallationPlatforms return the list of platforms
var InstallationPlatforms = map[dbaasv1alpha1.PlatformsName]dbaasv1alpha1.PlatformConfig{
	dbaasv1alpha1.DBaaSDynamicPluginInstallation: {
		Name:        dbaassDynamicPluginName,
		Image:       dbaassDynamicPluginImg,
		DisplayName: dbaassDynamicPluginDisplayName,
		Type:        dbaasv1alpha1.TypeConsolePlugin,
	},
	dbaasv1alpha1.ConsoleTelemetryPluginInstallation: {
		Name:        consoleTelemetryPluginName,
		Image:       consoleTelemetryPluginImg,
		DisplayName: consoleTelemetryPluginDisplayName,
		Envs:        []corev1.EnvVar{{Name: consoleTelemetryPluginSegmentKeyEnv, Value: consoleTelemetryPluginSegmentKey}},
		Type:        dbaasv1alpha1.TypeConsolePlugin,
	},
	dbaasv1alpha1.CrunchyBridgeInstallation: {
		Name:           crunchyBridgeName,
		CSV:            crunchyBridgeCSV,
		DeploymentName: crunchyBridgeDeployment,
		Image:          crunchyBridgeCatlaogImg,
		PackageName:    crunchyBridgePkg,
		Channel:        crunchyBridgeChannel,
		DisplayName:    crunchyBridgeDisplayName,
		Type:           dbaasv1alpha1.TypeProvider,
	},
	dbaasv1alpha1.MongoDBAtlasInstallation: {
		Name:           mongoDBAtlasName,
		CSV:            mongoDBAtlasCSV,
		DeploymentName: mongoDBAtlasDeployment,
		Image:          mongoDBAtlasCatalogImg,
		PackageName:    mongoDBAtlasPkg,
		Channel:        mongoDBAtlasChannel,
		DisplayName:    mongoDBAtlasDisplayName,
		Type:           dbaasv1alpha1.TypeProvider,
	},
	dbaasv1alpha1.CockroachDBInstallation: {
		Name:           cockroachDBName,
		CSV:            cockroachDBCSV,
		DeploymentName: cockroachDBDeployment,
		Image:          cockroachDBCatalogImg,
		PackageName:    cockroachDBPkg,
		Channel:        cockroachDBChannel,
		DisplayName:    cockroachDBDisplayName,
		Type:           dbaasv1alpha1.TypeProvider,
	},
	dbaasv1alpha1.DBaaSQuickStartInstallation: {
		Type: dbaasv1alpha1.TypeQuickStart,
	},
	dbaasv1alpha1.RDSProviderInstallation: {
		Name:           rdsProviderName,
		CSV:            rdsProviderCSV,
		DeploymentName: rdsProviderDeployment,
		Image:          rdsProviderCatalogImg,
		PackageName:    rdsProviderPkg,
		Channel:        rdsProviderChannel,
		DisplayName:    rdsProviderDisplayName,
		Type:           dbaasv1alpha1.TypeProvider,
	},
}
