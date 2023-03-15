/*
Copyright 2023 The OpenShift Database Access Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

// Constants for the providers supported
const (
	CockroachDBCloudRegistration = "cockroachdb-cloud-registration"
	MongoDBAtlasRegistration     = "mongodb-atlas-registration"
	CrunchyBridgeRegistration    = "crunchy-bridge-registration"
	RdsRegistration              = "rds-registration"
)

// Hub marks this type as a conversion hub.
func (*DBaaSProvider) Hub() {}
