# Database-as-a-Service Operator
The DBaaS Operator is a proof-of-concept golang kubernetes operator. The intent of the PoC is to show how we could
scan & import off-cluster cloud database instances hosted by various third-party providers & make those instances
available to developers for binding to their applications.

## Building the Operator
- `make build`

## Running the Operator
**NOTE**: The topology portion of the workflow described below will *only* work if your operator is installed via OLM.
An OLM-operator-backed resource is required for this bit to function. If you run locally or via direct deploy (first 2
options), you can create a DBaaSInventory & will receive a DBaaSConnection, but will not see the DBaaSConnection as
bindable in Topology view.

**Run as a local instance**:
- `make install run WATCH_NAMESPACE=<your_target_namespace>`
- Continue below by following the [Using the Operator](#using-the-operator) section
- When finished, remove created resources via:
  - `make clean-namespace`

**Deploy & run on a cluster:**
- `oc project <your_target_namespace>`
- `make deploy`
- Continue below by following the [Using the Operator](#using-the-operator) section
- When finished, clean up & remove deployment via:
  - `make clean-namespace undeploy`

**Deploy via OLM on cluster:**
- **Make sure to edit `Makefile` and replace `ORG` in the `IMAGE_TAG_BASE` with your own Quay.io Org!**
- **Next edit the [catalog-source.yaml](config/samples/catalog-source.yaml) template to indicate your new Quay.io org image**
- Edit the [catalog-operator-group.yaml](config/samples/catalog-operator-group.yaml) to indicate your target namespace
- `make install release catalog-update`
- `oc project <your_target_namespace>`
- `make deploy-olm`
- Continue below by following the [Using the Operator](#using-the-operator) section
- If you wish to uninstall operator from your project:
  - `make clean-namespace undeploy-olm`

## Using the Operator

**Prerequisites:**
- In this initial phase of PoC, we're working with MongoDB Atlas as our first cloud instances provider.
- The full workflow assumes that you're using a cluster with a modified MongoDB Atlas operator also running. The Atlas
  operator is used to fetch details about Atlas resources via API (with Organization public/private key-pair auth).
  - [MongoDB Atlas](https://github.com/RHEcosystemAppEng/mongodb-atlas-kubernetes)
  - [Crunchy Bridge PostgreSQL](https://github.com/CrunchyData/crunchy-bridge-operator)

**Creating a DBaaSInventory:**
- The DBaaSInventory resource indicates a request for the Atlas operator to fetch available resources. Eventually, we'll
  utilize the OpenShift console Administrator workflow to create & modify these resources, but for the sake of running
  locally or purely via CLI, you can create this resource as follows
- Once created, the `dbaas-operator` will reconcile this new DBaaSInventory resource & create an `MongoDBAtlasInventory` resource
  for the Atlas operator to consume.

**Reading resulting Atlas information:**
- Once the Atlas operator has completed its resource fetch of resources available in the cloud via your provided Atlas
  credentials, the dbaas-operator will read the response & update the DBaaSInventory `status` section with the resulting
  information. You can view an example of what a return status would look like in the
  [DBaaSInventory template](config/samples/dbaas_v1_DBaaSInventory.yaml).

**Specifying a cluster instance for import:**
- At this point in the workflow, the cluster instances available for import would be displayed to the OCP Admin for
  selecting which instances to import. We can mock this action manually as follows:
  - Pick a cluster instance ID from the DBaaSInventory `status` section to indicate for import
    - an example in the template status would be `6086aa1528fd4b1234aed133`
  - Update the DBaaSInventory `spec` to include the instance ID & save

**DBaaSConnection resource:**
- The `dbaas-operator` will read the import request and create four objects:
  - a ConfigMap containing the host URI and database selected for import
    - name: `dbaas-atlas-connection-<instance_id>`
  - a Secret containing the database user credentials for connecting
    - name: `dbaas-atlas-connection-<instance_id>`
  - a DBaaSConnection resource with a `spec` detailing the cluster instance information & `status` indicating references
    to the ConfigMap & Secret also created.
    - the DBaaSConnection CRD contains a series of annotations that mark the resource as bindable by the
      [Service Binding Operator](https://github.com/redhat-developer/service-binding-operator):
      ```
      service.binding/database: 'path={.status.dbConfigMap},objectType=ConfigMap'
      service.binding/host: 'path={.status.dbConfigMap},objectType=ConfigMap'
      service.binding/password: 'path={.status.dbCredentials},objectType=Secret'
      service.binding/provider: 'path={.spec.provider}'
      service.binding/type: 'path={.spec.type}'
      service.binding/username: 'path={.status.dbCredentials},objectType=Secret'
      ```
  - a zero-replica Deployment resource owned by the DBaaSConnection resource to make it viewable in the OCP Developer
    topology view as a bindable resource
    - NOTE: this deployment serves as a temporary PoC workaround to make the DBaaSConnection viewable, eventually it will
      become part of the developer workflow as a standalone component.

**Cleaning up:**
- Removing the sample binding & Quarkus application:
  - `make undeploy-sample-app`
- Remove DBaaS created resources:
  - `make clean-namespace`
- Remove the operator via whichever method you used for deployment (see above)

## Contributing

- Fork DBaas Operator repository
  - https://github.com/RHEcosystemAppEng/dbaas-operator
- Check out code from your new fork
  - `git clone git@github.com:<your-user-name>/dbaas-operator.git`
  - `cd dbaas-operator`
- Add upstream as git remote entry
  - `git remote add upstream git@github.com:RHEcosystemAppEng/dbaas-operator.git`
- create feature branches within your fork to complete your work
- raise PR's from your feature branch targeting upstream main branch
- add `jeremyary` (and others as needed) as reviewer