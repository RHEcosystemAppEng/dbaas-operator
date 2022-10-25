/*
Copyright 2022.

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

package v1alpha2

const (
	DBaaSServiceNotAvailable string = "DBaaSServiceNotAvailable"
)

// LocalObjectReference contains enough information to let you locate the
// referenced object inside the same namespace.
type LocalObjectReference struct {
	// Name of the referent.
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`
}

// NamespacedName defines the namespace and name of a k8s resource
type NamespacedName struct {
	// The namespace where object of known type is stored
	Namespace string `json:"namespace,omitempty"`

	// The name for object of known type
	Name string `json:"name"`
}

// DatabaseServiceType type of the supported database service
type DatabaseServiceType string

// Constants for database service types
const (
	InstanceDatabaseService DatabaseServiceType = "instance"
	ClusterDatabaseService  DatabaseServiceType = "cluster"
)
