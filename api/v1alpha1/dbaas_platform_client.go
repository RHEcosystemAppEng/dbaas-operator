package v1alpha1

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

type v1Alpha1Interface interface {
	DbaaSPlatform(namespace string) DbaaSPlatformInterface
}

type v1Alpha1Client struct {
	restClient rest.Interface
	ctx        context.Context
}

func NewForConfig(c *rest.Config) (*v1Alpha1Client, error) {
	config := *c
	config.ContentConfig.GroupVersion = &schema.GroupVersion{Group: GroupName, Version: Version}
	config.APIPath = "/apis"
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
	config.UserAgent = rest.DefaultKubernetesUserAgent()

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	return &v1Alpha1Client{restClient: client, ctx: ctx}, nil
}

func (c *v1Alpha1Client) DbaaSPlatform(namespace string) DbaaSPlatformInterface {
	return &dbaasPlatformClient{
		restClient: c.restClient,
		ns:         namespace,
		ctx:        c.ctx,
	}
}
