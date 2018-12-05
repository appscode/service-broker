package namespace

import (
	"fmt"
	"sync"

	hookapi "github.com/appscode/kubernetes-webhook-util/admission/v1beta1"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	admission "k8s.io/api/admission/v1beta1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

type NamespaceValidator struct {
	Resources   []string
	dc          dynamic.Interface
	lock        sync.RWMutex
	initialized bool
}

var _ hookapi.AdmissionHook = &NamespaceValidator{}

func (a *NamespaceValidator) Resource() (plural schema.GroupVersionResource, singular string) {
	return schema.GroupVersionResource{
			Group:    "validators.kubedb.com",
			Version:  "v1alpha1",
			Resource: "namespaces",
		},
		"namespace"
}

func (a *NamespaceValidator) Initialize(config *rest.Config, stopCh <-chan struct{}) error {
	a.lock.Lock()
	defer a.lock.Unlock()

	a.initialized = true

	var err error
	if a.dc, err = dynamic.NewForConfig(config); err != nil {
		return err
	}
	return err
}

func (a *NamespaceValidator) Admit(req *admission.AdmissionRequest) *admission.AdmissionResponse {
	status := &admission.AdmissionResponse{}

	// No validation on CREATE
	if (req.Operation != admission.Delete) ||
		len(req.SubResource) != 0 ||
		req.Kind.Group != core.SchemeGroupVersion.Group ||
		req.Kind.Kind != "Namespace" {
		status.Allowed = true
		return status
	}

	a.lock.RLock()
	defer a.lock.RUnlock()
	if !a.initialized {
		return hookapi.StatusUninitialized()
	}

	switch req.Operation {
	case admission.Delete:
		if req.Name != "" {
			var wg sync.WaitGroup

			results := make([]error, len(a.Resources))
			for idx, resource := range a.Resources {
				// Increment the WaitGroup counter.
				wg.Add(1)
				// Launch a goroutine to check a database kind.
				go func(idx int, resource string) {
					// Decrement the counter when the goroutine completes.
					defer wg.Done()

					list, err := a.dc.
						Resource(api.SchemeGroupVersion.WithResource(resource)).
						Namespace(req.Name).
						List(metav1.ListOptions{})
					if err != nil {
						results[idx] = err
						return
					}

					results[idx] = list.EachListItem(func(o runtime.Object) error {
						u := o.(*unstructured.Unstructured)
						doNotPause, found, err := unstructured.NestedBool(u.Object, "spec", "doNotPause")
						if err != nil {
							return err
						}
						if found && doNotPause {
							return fmt.Errorf("%s %s/%s can't be paused", u.GetKind(), u.GetNamespace(), u.GetName())
						}

						terminationPolicy, found, err := unstructured.NestedString(u.Object, "spec", "terminationPolicy")
						if err != nil {
							return err
						}
						if !found ||
							terminationPolicy == string(api.TerminationPolicyPause) ||
							terminationPolicy == string(api.TerminationPolicyDoNotTerminate) {
							return fmt.Errorf("%s %s/%s has termination policy `%s`", u.GetKind(), u.GetNamespace(), u.GetName(), terminationPolicy)
						}
						return nil
					})
				}(idx, resource)
			}
			// Wait for all checks to complete.
			wg.Wait()

			if err := utilerrors.NewAggregate(results); err != nil {
				return hookapi.StatusForbidden(err)
			}
		}
	}

	status.Allowed = true
	return status
}
