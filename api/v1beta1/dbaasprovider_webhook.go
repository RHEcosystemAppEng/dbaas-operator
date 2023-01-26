/*
Copyright 2021.

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

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// log is for logging in this package.
var dbaasproviderlog = logf.Log.WithName("dbaasprovider-resource")

func (r *DBaaSProvider) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

func (r *DBaaSProvider) GetDBaaSAPIGroupVersion() schema.GroupVersion {
	if len(r.Spec.GroupVersion) > 0 {
		groupVersion, err := schema.ParseGroupVersion(r.Spec.GroupVersion)
		if err == nil {
			return groupVersion
		}
		// If the group version is not valid, default to v1alpha1
	}
	// Use v1alpha1 if API version is not defined (for backward compatibility)
	return schema.GroupVersion{
		Group:   GroupVersion.Group,
		Version: "v1alpha1",
	}
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
