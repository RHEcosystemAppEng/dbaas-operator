
## How do we integrate DBaaS Operator with [Observability Operator](https://github.com/rhobs/observability-operator)

**Note: This is only necessary if you intend to configure Observatorium for the DBaaS operator and send the metrics to the central Observatorium. Otherwise, this step is not required**


### Observability Operator Remote Write Configuration

The monitoring stack is CR for the observability operator created from the DBaaS operator automatically that will create required resources for monitoring for us in the same namespace.


For configuring the Monitoring Stack CR for remote writing into the Observatorium. We have a PrometheusConfig field under MonitoringStack CR that enables us to configure the Prometheus for Remote writing.
This all done by DBaaS operator within the [code](https://github.com/RHEcosystemAppEng/dbaas-operator/blob/main/controllers/reconcilers/providersinstallation/observability-operator.go). However, you need to deploy DBaaS operator, by defining additional below parameters under operator subscription environment variables, and a secret which contains the Observatorium token or SSO configuration depend on RHOBS authentication type.

###### Apply the Secret with token, substituted in:

```oc apply -f config/samples/secret-dbaas-operator-dev-prom-remote-write.yaml```

The following parameters can be modified or configured in operator subscription environment variables.

      - name: ADDON_NAME
        value: dbaas-operator-dev or dbaas-operator-stage or dbaas-operator-prod
      - name: RHOBS_AUTH_TYPE
        value: 'dex' if token, else `redhat-sso` if using SSO for authentication
      - name: RHOBS_API_URL
        value: https://observatorium-observatorium.apps.<>/api/metrics/v1/test2/api/v1/receive

###### Install the RHODA operator, by substituted above params in:

```oc apply -f config/samples/installation-for-non-OSD-OCP-environments.yaml```
