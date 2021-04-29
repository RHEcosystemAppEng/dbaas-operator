# DBaaS Operator


### Contributing

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

**Building the operator**
  
- Update the `Makefile` and edit `IMAGE_TAG_BASE`, with your `ORG` name.

- To build and push the operator, bundle and index images, you just need to run: `make`

**Run locally outside the cluster**:  `make install run`
 
**Deploy manualy inside the cluster**: Run `make deploy` 
 
**Deploy With OLM**: Run `operator-sdk run bundle quay.io/ecosystem-appeng/dbaas-operator-bundle:<TAG>`
   
**Create the CR**:  `oc apply -f config/samples/dbaas_v1_dbaasservice.yaml -n <namesapce> `

**Cleanup**:  

  `oc delete -f config/samples/dbaas_v1_dbaasservice.yaml -n <namesapce> `
  
  `make undeploy`
  
  cleanup from OLM: `operator-sdk cleanup dbaas-operator`
