#! /bin/bash
echo "This script will completely remove RHODA, Providers, and all associated CRD's from a cluster.
It should be considered a deep-cleaning as all usages & CRD's will be removed.

This script will:
1) delete all user-created DBaaS Provider Accounts, Instances, and Connections
2) delete our DBaas Platform resource to trigger removal of provider operators
3) delete all DBaaS Tenant resources as they\'ll no longer be required
4) delete any orphaned provider resources just in case some user managed an edge case
5) uninstall the RHODA operator via deleting the Subscription, OperatorGroup & CatalogSource
6) delete all CRD\'s associated to the RHODA and provider operators

You must have logged into the OpenShift cluster as a cluster administrator in order to proceed."
read -p "Would you like to continue (y/n)?" answer
case ${answer:0:1} in
    y|Y )
        echo Yes
    ;;
    * )
        exit 1
    ;;
esac

# Remove any of our custom resources in use
echo "removing DBaas specific resources..."
oc delete dbaasconnections,dbaasinstances,dbaasinventories --all-namespaces --all

sleep 5

# Next, delete Platform CR to trigger dependency removal
echo "removing platform resource..."
oc delete dbaasplatforms -n redhat-dbaas-operator --all

# No longer required, remove all Tenant resources
echo "removing tenant resources..."
oc delete dbaaspolicies  --all-namespaces --all

sleep 5

# take care of any stubborn provider resources that survived the dbaas resource removal just in case some user orphaned something
echo "removing edge case provider resources..."
oc delete mongodbatlasconnections,mongodbatlasinstances,mongodbatlasinventories,crdbdbaasconnections,crdbdbaasinstances,crdbdbaasinventories,crunchybridgeconnections,crunchybridgeinstances,crunchybridgeinventories --all-namespaces --all --grace-period=0 --force

# uninstall operator
echo "removing OLM resources..."
oc delete suscription dbaas-sub -n redhat-dbaas-operator
oc delete operatorgroup dbaas-operator -n redhat-dbaas-operator
oc delete catalogsource -n openshift-marketplace dbaas-operator-catalog crunchy-bridge-catalogsource ccapi-k8s-catalogsource mongodb-atlas-catalogsource rds-provider-catalogsource

# remove all DBaaS and Provider CRD's
echo "removing CRD's..."
oc delete crd dbaasinventories.dbaas.redhat.com dbaasconnections.dbaas.redhat.com dbaaspolicies.dbaas.redhat.com dbaasinstances.dbaas.redhat.com dbaasplatforms.dbaas.redhat.com dbaasproviders.dbaas.redhat.com crdbdbaasconnections.dbaas.redhat.com crdbdbaasinstances.dbaas.redhat.com crdbdbaasinventories.dbaas.redhat.com crunchybridgeconnections.dbaas.redhat.com crunchybridgeinventories.dbaas.redhat.com crunchybridgeinstances.dbaas.redhat.com mongodbatlasconnections.dbaas.redhat.com mongodbatlasinventories.dbaas.redhat.com mongodbatlasinstances.dbaas.redhat.com rdsconnections.dbaas.redhat.com rdsinventories.dbaas.redhat.com rdsinstances.dbaas.redhat.com

echo "all done!"
exit 1