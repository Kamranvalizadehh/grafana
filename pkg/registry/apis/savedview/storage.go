package savedview

import (
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/generic"
	genericregistry "k8s.io/apiserver/pkg/registry/generic/registry"

	savedview "github.com/grafana/grafana/pkg/apis/savedview/v0alpha1"
	grafanaregistry "github.com/grafana/grafana/pkg/apiserver/registry/generic"
	grafanarest "github.com/grafana/grafana/pkg/apiserver/rest"
	gapiutil "github.com/grafana/grafana/pkg/services/apiserver/utils"
)

var _ grafanarest.Storage = (*storage)(nil)

type storage struct {
	*genericregistry.Store
}

func newStorage(scheme *runtime.Scheme, optsGetter generic.RESTOptionsGetter) (*storage, error) {
	strategy := grafanaregistry.NewStrategy(scheme)

	resourceInfo := savedview.SavedViewResourceInfo
	store := &genericregistry.Store{
		NewFunc:                   resourceInfo.NewFunc,
		NewListFunc:               resourceInfo.NewListFunc,
		KeyRootFunc:               grafanaregistry.KeyRootFunc(resourceInfo.GroupResource()),
		KeyFunc:                   grafanaregistry.NamespaceKeyFunc(resourceInfo.GroupResource()),
		PredicateFunc:             grafanaregistry.Matcher,
		DefaultQualifiedResource:  resourceInfo.GroupResource(),
		SingularQualifiedResource: resourceInfo.SingularGroupResource(),
		TableConvertor: gapiutil.NewTableConverter(
			resourceInfo.GroupResource(),
			[]metav1.TableColumnDefinition{
				{Name: "Name", Type: "string", Format: "name"},
				{Name: "Url", Type: "string", Format: "string", Description: "The url"},
				{Name: "Created At", Type: "date"},
			},
			func(obj any) ([]interface{}, error) {
				m, ok := obj.(*savedview.SavedView)
				if !ok {
					return nil, fmt.Errorf("expected savedview")
				}
				return []interface{}{
					m.Name,
					m.Spec.URL,
					m.CreationTimestamp.UTC().Format(time.RFC3339),
				}, nil
			},
		),
		CreateStrategy: strategy,
		UpdateStrategy: strategy,
		DeleteStrategy: strategy,
	}
	options := &generic.StoreOptions{RESTOptions: optsGetter, AttrFunc: grafanaregistry.GetAttrs}
	if err := store.CompleteWithOptions(options); err != nil {
		return nil, err
	}
	return &storage{Store: store}, nil
}