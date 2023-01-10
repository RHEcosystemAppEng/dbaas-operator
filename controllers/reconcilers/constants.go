package reconcilers

import (
	"fmt"
	"os"

	dbaasv1beta1 "github.com/RHEcosystemAppEng/dbaas-operator/api/v1beta1"
	embeddedconfigs "github.com/RHEcosystemAppEng/dbaas-operator/config"
	"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	// InstallNamespace namespace for installation
	InstallNamespace = "openshift-operators"
	// CatalogNamespace namespace for catalog sources
	CatalogNamespace = "openshift-marketplace"

	// DBaaSQuickStartVersion version for the quick start guide
	DBaaSQuickStartVersion = "dbaas-quick-starts:0.4.0"

	// CRUNCHY_BRIDGE
	crunchyBridgeCatalogImg  = "RELATED_IMAGE_CRUNCHY_BRIDGE_CATALOG"
	crunchyBridgeCSV         = "CSV_VERSION_CRUNCHY_BRIDGE"
	crunchyBridgeName        = "crunchy-bridge"
	crunchyBridgeDisplayName = "Crunchy Bridge Operator"
	crunchyBridgeDeployment  = "crunchy-bridge-operator-controller-manager"
	crunchyBridgePkg         = "crunchy-bridge-operator"
	crunchyBridgeChannel     = "alpha"

	// MONGODB_ATLAS
	mongoDBAtlasCatalogImg  = "RELATED_IMAGE_MONGODB_ATLAS_CATALOG"
	mongoDBAtlasCSV         = "CSV_VERSION_MONGODB_ATLAS"
	mongoDBAtlasName        = "mongodb-atlas"
	mongoDBAtlasDisplayName = "MongoDB Atlas Operator"
	mongoDBAtlasDeployment  = "mongodb-atlas-operator"
	mongoDBAtlasPkg         = "mongodb-atlas-kubernetes"
	mongoDBAtlasChannel     = "beta"

	// COCKROACHDB
	cockroachDBCSV         = "CSV_VERSION_COCKROACHDB"
	cockroachDBCatalogImg  = "RELATED_IMAGE_COCKROACHDB_CATALOG"
	cockroachDBName        = "ccapi-k8s"
	cockroachDBDisplayName = "CockroachDB Cloud Operator"
	cockroachDBDeployment  = "ccapi-k8s-operator-controller-manager"
	cockroachDBPkg         = "ccapi-k8s-operator"
	cockroachDBChannel     = "alpha"

	// DBAAS_DYNAMIC_PLUGIN
	dbaasDynamicPluginImg         = "RELATED_IMAGE_DBAAS_DYNAMIC_PLUGIN"
	dbaasDynamicPluginVersion     = "CSV_VERSION_DBAAS_DYNAMIC_PLUGIN"
	dbaasDynamicPluginName        = "dbaas-dynamic-plugin"
	dbaasDynamicPluginDisplayName = "OpenShift Database as a Service Dynamic Plugin"

	// RDS_PROVIDER
	rdsProviderCSV         = "CSV_VERSION_RDS_PROVIDER"
	rdsProviderCatalogImg  = "RELATED_IMAGE_RDS_PROVIDER_CATALOG"
	rdsProviderName        = "rds-provider"
	rdsProviderDisplayName = "RHODA Provider Operator for Amazon RDS"
	rdsProviderDeployment  = "rds-dbaas-operator-controller-manager"
	rdsProviderPkg         = "rds-dbaas-operator"
	rdsProviderChannel     = "alpha"

	//ObservabilityName platform name for observability
	ObservabilityName = "observability"
	//other constants for observability
	observabilityCatalogImg  = "RELATED_IMAGE_OBSERVABILITY_CATALOG"
	observabilityCSV         = "CSV_VERSION_OBSERVABILITY"
	observabilityDisplayName = "observability Operator"
	observabilityDeployment  = "observability-operator"
	observabilityPkg         = "observability-operator"
	observabilityChannel     = "stable"
)

// InstallationPlatforms return the list of platforms
var InstallationPlatforms = map[dbaasv1beta1.PlatformName]dbaasv1beta1.PlatformConfig{
	dbaasv1beta1.DBaaSDynamicPluginInstallation: {
		Name:        dbaasDynamicPluginName,
		CSV:         fetchEnvValue(dbaasDynamicPluginVersion),
		Image:       fetchEnvValue(dbaasDynamicPluginImg),
		DisplayName: dbaasDynamicPluginDisplayName,
		Type:        dbaasv1beta1.TypeConsolePlugin,
	},
	dbaasv1beta1.CrunchyBridgeInstallation: {
		Name:           crunchyBridgeName,
		CSV:            fetchEnvValue(crunchyBridgeCSV),
		DeploymentName: crunchyBridgeDeployment,
		Image:          fetchEnvValue(crunchyBridgeCatalogImg),
		PackageName:    crunchyBridgePkg,
		Channel:        crunchyBridgeChannel,
		DisplayName:    crunchyBridgeDisplayName,
		Type:           dbaasv1beta1.TypeOperator,
	},
	dbaasv1beta1.MongoDBAtlasInstallation: {
		Name:           mongoDBAtlasName,
		CSV:            fetchEnvValue(mongoDBAtlasCSV),
		DeploymentName: mongoDBAtlasDeployment,
		Image:          fetchEnvValue(mongoDBAtlasCatalogImg),
		PackageName:    mongoDBAtlasPkg,
		Channel:        mongoDBAtlasChannel,
		DisplayName:    mongoDBAtlasDisplayName,
		Type:           dbaasv1beta1.TypeOperator,
	},
	dbaasv1beta1.CockroachDBInstallation: {
		Name:           cockroachDBName,
		CSV:            fetchEnvValue(cockroachDBCSV),
		DeploymentName: cockroachDBDeployment,
		Image:          fetchEnvValue(cockroachDBCatalogImg),
		PackageName:    cockroachDBPkg,
		Channel:        cockroachDBChannel,
		DisplayName:    cockroachDBDisplayName,
		Type:           dbaasv1beta1.TypeOperator,
	},
	dbaasv1beta1.DBaaSQuickStartInstallation: {
		Type: dbaasv1beta1.TypeQuickStart,
		CSV:  DBaaSQuickStartVersion,
	},
	dbaasv1beta1.RDSProviderInstallation: {
		Name:           rdsProviderName,
		CSV:            fetchEnvValue(rdsProviderCSV),
		DeploymentName: rdsProviderDeployment,
		Image:          fetchEnvValue(rdsProviderCatalogImg),
		PackageName:    rdsProviderPkg,
		Channel:        rdsProviderChannel,
		DisplayName:    rdsProviderDisplayName,
		Type:           dbaasv1beta1.TypeOperator,
	},
	dbaasv1beta1.ObservabilityInstallation: {
		Name:           ObservabilityName,
		CSV:            fetchEnvValue(observabilityCSV),
		DeploymentName: observabilityDeployment,
		Image:          fetchEnvValue(observabilityCatalogImg),
		PackageName:    observabilityPkg,
		Channel:        observabilityChannel,
		DisplayName:    observabilityDisplayName,
		Type:           dbaasv1beta1.TypeOperator,
	},
}

// GetObservabilityConfig return observatorium configuration
func GetObservabilityConfig() dbaasv1beta1.ObservabilityConfig {
	return dbaasv1beta1.ObservabilityConfig{
		AuthType:        os.Getenv("RHOBS_AUTH_TYPE"),
		RemoteWritesURL: os.Getenv("RHOBS_API_URL"),
		RHSSOTokenURL:   os.Getenv("RH_SSO_TOKEN_ENDPOINT"),
		AddonName:       os.Getenv("ADDON_NAME"),
		RHOBSSecretName: fmt.Sprintf("%v-prom-remote-write", os.Getenv("ADDON_NAME")),
	}
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
