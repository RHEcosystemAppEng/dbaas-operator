package controllers

import (
	"go.uber.org/zap"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
)

// EventHandlerWithDelete is an extension of EnqueueRequestForObject that will _not_ trigger a reconciliation for a Delete event.
// Instead, it will call an external controller's Delete() method and pass the event argument unchanged.
type EventHandlerWithDelete struct {
	handler.EnqueueRequestForObject
	Controller interface {
		Delete(e event.DeleteEvent) error
	}
}

// Delete implements a handler for the Delete event.
func (d *EventHandlerWithDelete) Delete(e event.DeleteEvent, _ workqueue.RateLimitingInterface) {
	objectKey := objectKeyFromObject(e.Object)
	log := zap.S().With("resource", objectKey)

	if err := d.Controller.Delete(e); err != nil && k8serrors.IsNotFound(err) {
		log.Errorf("Object (%s) removed from Kubernetes, but controller could not delete it: %s", e.Object.GetObjectKind(), err)
	}
}

func objectKeyFromObject(obj metav1.Object) client.ObjectKey {
	return objectKey(obj.GetNamespace(), obj.GetName())
}

func objectKey(namespace, name string) client.ObjectKey {
	return types.NamespacedName{Name: name, Namespace: namespace}
}
