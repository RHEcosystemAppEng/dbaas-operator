# Install and Deploy MSO instance
___

We have to deploy the MSO after installing the dbaas-operator on the cluster so that we can see and query the metrics from Prometheus UI.

## Steps to Deploy MSO

```commandline
make install-mso 
// Deploys the catalog source and subscription for MSO.

make deploy-mso
// Applys the MSO CR which created the Prometheus instance.
```

## Steps to Remove MSO

```commandline
make undeploy-mso 
// Deletes the MSO CR.

make uninstall-mso
// Deletes the catalog source and subscription for MSO.
```