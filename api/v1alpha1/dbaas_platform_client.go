package v1alpha1

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

type v1Alpha1Interface interface {
	DbaaSPlatform(namespace string, resource string) DbaaSPlatformInterface
}

type v1Alpha1Client struct {
	restClient rest.Interface
	ctx        context.Context
}

func NewForConfig(c *rest.Config) (*v1Alpha1Client, error) {
	config := *c

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	return &v1Alpha1Client{restClient: client, ctx: ctx}, nil
}

func (c *v1Alpha1Client) DbaaSPlatform(namespace string, resource string) DbaaSPlatformInterface {
	return &dbaasPlatformClient{
		restClient: c.restClient,
		ns:         namespace,
		ctx:        c.ctx,
		resource:   resource,
	}
}

//+kubebuilder:object:generate=false

type DbaaSPlatformInterface interface {
	List(opts metav1.ListOptions) (*DBaaSPlatformList, error)
	Get(name string, options metav1.GetOptions) (*DBaaSPlatform, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
}

type dbaasPlatformClient struct {
	restClient rest.Interface
	ns         string
	ctx        context.Context
	resource   string
}

func (c *dbaasPlatformClient) List(opts metav1.ListOptions) (*DBaaSPlatformList, error) {
	result := DBaaSPlatformList{}
	err := c.restClient.
		Get().
		Namespace(c.ns).
		Resource(c.resource).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(c.ctx).
		Into(&result)

	return &result, err
}

func (c *dbaasPlatformClient) Get(name string, opts metav1.GetOptions) (*DBaaSPlatform, error) {
	result := DBaaSPlatform{}
	err := c.restClient.
		Get().
		Namespace(c.ns).
		Resource(c.resource).
		Name(name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(c.ctx).
		Into(&result)

	return &result, err
}

func (c *dbaasPlatformClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.restClient.
		Get().
		Namespace(c.ns).
		Resource(c.resource).
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch(c.ctx)
}
