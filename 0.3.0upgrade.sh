#! /bin/bash
# this should be run if RHODA gets stuck in "pending" when attempting to upgrade from 0.2.0 -> 0.3.0
set -e

if ! which oc > /dev/null; then
   echo "'oc' command not found"
   echo 'install and try again - https://docs.okd.io/latest/cli_reference/openshift_cli/getting-started-cli.html'
   exit
fi

# oc login ... (as user with admin access to the dbaas operator install ns. e.g. openshift-dbaas-operator / redhat-dbaas-operator)

# first verify 0.2.0 has previously been deployed
installns02=$(oc get csv dbaas-operator.v0.2.0 --ignore-not-found -o template --template '{{index .metadata.annotations "olm.operatorNamespace"}}')
installns03=$(oc get csv dbaas-operator.v0.3.0 --ignore-not-found -o template --template '{{index .metadata.annotations "olm.operatorNamespace"}}')

if [ ! -z ${installns02} ] && [ ! -z ${installns03} ]; then
    echo ""
    echo ${installns02}

    subname=$(oc get sub addon-dbaas-operator -n ${installns02} --ignore-not-found --template '{{.metadata.name}}')
    if [ -z ${subname} ]; then
        subname=$(oc get sub dbaas-operator -n ${installns02} --ignore-not-found --template '{{.metadata.name}}')
    fi

    if [ ! -z ${subname} ]; then
        # add if check to see if manager exists first
        deploy=$(oc get deploy dbaas-operator-controller-manager -n ${installns02} --ignore-not-found --template '{{.metadata.name}}')
        if [ ! -z ${deploy} ]
        then
            # stop 0.2.0 operator
            oc scale --replicas=0 deploy dbaas-operator-controller-manager -n ${installns02}
            sleep 3
        fi

        oc patch sub ${subname} -n ${installns02} --type=merge -p '{"spec":{"startingCSV": "dbaas-operator.v0.3.0"}}'
        subspec=$(oc get sub ${subname} -n ${installns02} -o jsonpath='{.spec}')

        subscription=$(cat <<EOF
{
    "apiVersion":"operators.coreos.com/v1alpha1",
    "kind":"Subscription",
    "metadata":{
        "name":"${subname}",
        "namespace":"${installns02}"
    },
    "spec":${subspec}
}
EOF
)
        echo ""
        echo "Subscription which will be applied by this script after some further cleanup."
        echo ${subscription}
        echo ""

        oc delete sub \
            ack-rds-controller-alpha-community-operators-openshift-marketplace \
            ${subname} \
            --ignore-not-found -n ${installns02}
        sleep 5

        oc delete dbaasplatform dbaas-platform --ignore-not-found -n ${installns02}
        sleep 3

        # upgrade should succeed regardless, but will attempt to remove the offending crd
        if [ $(oc auth can-i delete crds) == "yes" ]; then
            oc delete crd dbaasplatforms.dbaas.redhat.com --ignore-not-found
        fi

        oc delete csv \
            ack-rds-controller.v0.0.27 \
            ccapi-k8s-operator.v0.0.3 \
            crunchy-bridge-operator.v0.0.5 \
            rds-dbaas-operator.v0.1.0 \
            mongodb-atlas-kubernetes.v0.2.0 \
            dbaas-operator.v0.2.0 \
            dbaas-operator.v0.3.0 \
            --ignore-not-found -n ${installns02}

        cat <<EOF | oc create -f -
${subscription}
EOF
    fi
else
    echo "Nothing to do"
fi
