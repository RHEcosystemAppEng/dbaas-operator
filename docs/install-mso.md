# Install and Deploy Observability Operator instance
___

We are installing Observability operator like other Provider operators. After the Dbaas operator is deployed, we apply the CR for Monitoring Stack.

### Steps to Deploy MSO CR

```commandline
oc project openshift-dbaas-operator
//select the namespace you deployed the dbaas-operator in

make deploy-mso
// Deploys the CR.
```

### Steps to Remove MSO CR

```commandline
make undeploy-mso 
// Deletes the MSO CR.

```