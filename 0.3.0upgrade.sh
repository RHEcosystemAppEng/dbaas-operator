#! /bin/bash
# this should be run if RHODA gets stuck in "pending" when attempting to upgrade from 0.2.0 -> 0.3.0
set -e

if ! which oc > /dev/null; then
   echo "'oc' command not found"
   echo 'install and try again - https://docs.okd.io/latest/cli_reference/openshift_cli/getting-started-cli.html'
   exit
fi

# oc login ... (as user with admin access to the dbaas operator install ns. e.g. openshift-dbaas-operator / redhat-dbaas-operator)
ocuser=$(oc whoami)
echo "Logged in as ${ocuser}"

if [ $(oc auth can-i get csv) != "yes" ]; then
    oc project
    exit
fi

# first verify 0.2.0 has previously been deployed
installns02=$(oc get csv dbaas-operator.v0.2.0 --ignore-not-found -o template --template '{{index .metadata.annotations "olm.operatorNamespace"}}')
installns03=$(oc get csv dbaas-operator.v0.3.0 --ignore-not-found -o template --template '{{index .metadata.annotations "olm.operatorNamespace"}}')

if [ ! -z ${installns02} ] && [ ! -z ${installns03} ]; then
    echo ""
    echo "Running script against ${installns02} project"

    if [ $(oc auth can-i update roles -n ${installns02}) != "yes" ]; then
        echo "'oc login ...' with a user that has admin rights to the ${installns02} project and try again"
        exit
    fi

    subname=$(oc get sub addon-dbaas-operator -n ${installns02} --ignore-not-found --template '{{.metadata.name}}')
    if [ -z ${subname} ]; then
        subname=$(oc get sub dbaas-operator -n ${installns02} --ignore-not-found --template '{{.metadata.name}}')
    fi

    if [ ! -z ${subname} ]; then
        # add if check to see if manager exists first
        deploy=$(oc get deploy dbaas-operator-controller-manager -n ${installns02} --ignore-not-found --template '{{.metadata.name}}')
        if [ ! -z ${deploy} ]; then
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
        echo "If the script fails mid-run for an unexpected reason, you will need to apply this subscription manually."
        echo "PLEASE COPY THE SUBSCRIPTION JSON ABOVE TO A SAFE PLACE BEFORE CONTINUING."
        echo ""
        printf "Press any key to continue or 'CTRL+C' to exit:\n"

        # preserve input mode for restoration
        (tty_state=$(stty -g)
        # swap to canonical input to respect signals and allow ctrl+c
        stty -icanon
        # take input
        LC_ALL=C dd bs=1 count=1 >/dev/null 2>&1
        #restore input mode
        stty "$tty_state"
        ) </dev/tty

        oc delete sub \
            ack-rds-controller-alpha-community-operators-openshift-marketplace \
            ${subname} \
            --ignore-not-found -n ${installns02}
        sleep 5

        oc delete dbaasplatform dbaas-platform --ignore-not-found -n ${installns02}
        sleep 3

        # upgrade should succeed regardless, but will attempt to remove the offending crd
        if [ $(oc auth can-i delete crds --all-namespaces) == "yes" ]; then
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
