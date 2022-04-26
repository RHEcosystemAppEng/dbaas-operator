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

package v1alpha1

import (
	"context"
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	k8s "sigs.k8s.io/controller-runtime/pkg/client"

	vault "github.com/hashicorp/vault/api"
	auth "github.com/hashicorp/vault/api/auth/kubernetes"
)

//+kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list

// Fetches a key-value secret (kv-v2) after authenticating to Vault with a Kubernetes service account.
func GetSecretWithKubernetesAuth(k8sClient k8s.Client, role, path, sa, namespace string) (map[string]interface{}, error) {
	// If set, the VAULT_ADDR environment variable will be the address that
	// your pod uses to communicate with Vault.
	config := vault.DefaultConfig() // modify for more granular configuration

	client, err := vault.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize Vault client: %w", err)
	}

	var serviceAccount v1.ServiceAccount
	if err := k8sClient.Get(context.Background(),
		types.NamespacedName{Name: sa, Namespace: namespace},
		&serviceAccount); err != nil {
		return nil, err
	}
	var saSecret v1.Secret
	var saName string
	for _, s := range serviceAccount.Secrets {
		if strings.Contains(s.Name, "-token-") {
			saName = s.Name
			break
		}
	}
	if err := k8sClient.Get(context.Background(),
		types.NamespacedName{Name: saName, Namespace: namespace},
		&saSecret); err != nil {
		return nil, err
	}
	token, ok := saSecret.Data["token"]
	if !ok {
		return nil, fmt.Errorf("invalid token in secret")
	}
	k8sAuth, err := auth.NewKubernetesAuth(
		role,
		auth.WithServiceAccountToken(string(token)),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize Kubernetes auth method: %w", err)
	}

	authInfo, err := client.Auth().Login(context.TODO(), k8sAuth)
	if err != nil {
		return nil, fmt.Errorf("unable to log in with Kubernetes auth: %w", err)
	}
	if authInfo == nil {
		return nil, fmt.Errorf("no auth info was returned after login")
	}

	secret, err := client.Logical().Read(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read secret from path %v: %w", path, err)
	}

	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("data type assertion failed: %T %#v", secret.Data["data"], secret.Data["data"])
	}
	return data, nil
}
