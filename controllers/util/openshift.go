package util

import (
	"context"

	v1 "github.com/openshift/api/config/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// constants
const (
	VersionKeyName      = "version"
	ClusterResourceName = "cluster"
)

// GetClusterIDVersion Returns the cluster id and cluster version by querying the ClusterVersion resource
func GetClusterIDVersion(ctx context.Context, client k8sclient.Client) (string, string, error) {
	v := &v1.ClusterVersion{}
	selector := k8sclient.ObjectKey{
		Name: VersionKeyName,
	}

	err := client.Get(ctx, selector, v)
	if err != nil {
		return "", "", err
	}

	return string(v.Spec.ClusterID), v.Status.Desired.Version, nil
}

// GetOpenshiftConsoleURL Returns the openshift console URL
func GetOpenshiftConsoleURL(ctx context.Context, client k8sclient.Client) (string, error) {

	console := &v1.Console{}
	selector := k8sclient.ObjectKey{
		Name: ClusterResourceName,
	}
	err := client.Get(ctx, selector, console)
	if err != nil {
		return "", err
	}
	return console.Status.ConsoleURL, nil
}

// GetOpenshiftPlatform Returns the openshift deployment platform
func GetOpenshiftPlatform(ctx context.Context, client k8sclient.Client) (v1.PlatformType, error) {
	infra := &v1.Infrastructure{}
	selector := k8sclient.ObjectKey{
		Name: ClusterResourceName,
	}
	if err := client.Get(ctx, selector, infra); err != nil {
		if errors.IsNotFound(err) {
			return "", nil
		}
		return "", err
	}

	if infra.Status.PlatformStatus == nil {
		return "", nil
	}

	return infra.Status.PlatformStatus.Type, nil
}
