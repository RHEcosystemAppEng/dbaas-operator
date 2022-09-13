
## How do we integrate RHODA Operator with [Observability Operator](https://github.com/rhobs/observability-operator)

Observability Operator will be installed similar way as we install the other provider's operators/components. The Observability operator will be responsible to install the Prometheus, using the operator-owned deployment in the same namespace where the user installs the RHODA operator.

The monitoring stack is CR for the observability operator created from the RHODA operator automatically that will create required resources for monitoring for us in the same namespace.

### Observability Operator Remote Write Configuration

For configuring the Monitoring Stack CR for remote writing into the Observatorium. We have a PrometheusConfig field under MonitoringStack CR that enables us to configure the Prometheus for Remote writing.
This all done by RHODA operator within the [code](https://github.com/RHEcosystemAppEng/dbaas-operator/blob/main/controllers/reconcilers/providersinstallation/observability-operator.go). However, you need to deploy RHODA operator, by defining additional below parameters under operator subscription environment variables, and a secret which contains the Observatorium token or SSO configuration depend on RHOBS authentication type.

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
