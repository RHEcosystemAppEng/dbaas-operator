package v1alpha1

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

type v1Alpha1InventoryInterface interface {
	DbaaSInventory(namespace string) DbaaSInventoryInterface
}

type v1Alpha1InventoryClient struct {
	restClient rest.Interface
	ctx        context.Context
}

func NewConfigForInventory(c *rest.Config) (*v1Alpha1InventoryClient, error) {
	config := *c

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	return &v1Alpha1InventoryClient{restClient: client, ctx: ctx}, nil
}

func (c *v1Alpha1InventoryClient) DbaaSInventory(namespace string, resource string) DbaaSInventoryInterface {
	return &dbaaSInventoryClient{
		restClient: c.restClient,
		ns:         namespace,
		ctx:        c.ctx,
		resource:   resource,
	}
}

//+kubebuilder:object:generate=false

type DbaaSInventoryInterface interface {
	List(opts metav1.ListOptions) (*DBaaSInventoryList, error)
	Get(name string, options metav1.GetOptions) (*DBaaSInventory, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
}

type dbaaSInventoryClient struct {
	restClient rest.Interface
	ns         string
	ctx        context.Context
	resource   string
}

func (c *dbaaSInventoryClient) List(opts metav1.ListOptions) (*DBaaSInventoryList, error) {
	result := DBaaSInventoryList{}
	err := c.restClient.
		Get().
		Namespace(c.ns).
		Resource(c.resource).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(c.ctx).
		Into(&result)

	return &result, err
}

func (c *dbaaSInventoryClient) Get(name string, opts metav1.GetOptions) (*DBaaSInventory, error) {
	result := DBaaSInventory{}
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

func (c *dbaaSInventoryClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.restClient.
		Get().
		Namespace(c.ns).
		Resource(c.resource).
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch(c.ctx)
}
