package dashboard

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/generic"
	genericregistry "k8s.io/apiserver/pkg/registry/generic/registry"
	"k8s.io/apiserver/pkg/registry/rest"

	common "github.com/grafana/grafana/pkg/apimachinery/apis/common/v0alpha1"
	grafanaregistry "github.com/grafana/grafana/pkg/apiserver/registry/generic"
	"github.com/grafana/grafana/pkg/registry/apis/dashboard/access"
	"github.com/grafana/grafana/pkg/storage/unified/apistore"
	"github.com/grafana/grafana/pkg/storage/unified/resource"
)

type dashboardStorage struct {
	resource       common.ResourceInfo
	access         access.DashboardAccess
	tableConverter rest.TableConvertor
}

func (s *dashboardStorage) newStore(scheme *runtime.Scheme, defaultOptsGetter generic.RESTOptionsGetter) (rest.Storage, error) {
	server, err := resource.NewResourceServer(resource.ResourceServerOptions{
		Store: s.access,
		Blob:  s.access,
		// WriteAccess: resource.WriteAccessHooks{
		// 	Folder: func(ctx context.Context, user identity.Requester, uid string) bool {
		// 		// ???
		// 	},
		// },
	})
	if err != nil {
		return nil, err
	}

	resourceInfo := s.resource
	defaultOpts, err := defaultOptsGetter.GetRESTOptions(resourceInfo.GroupResource())
	if err != nil {
		return nil, err
	}
	client := resource.NewLocalResourceStoreClient(server)
	optsGetter := apistore.NewRESTOptionsGetter(client,
		defaultOpts.StorageConfig.Codec,
	)

	strategy := grafanaregistry.NewStrategy(scheme)
	store := &genericregistry.Store{
		NewFunc:                   resourceInfo.NewFunc,
		NewListFunc:               resourceInfo.NewListFunc,
		KeyRootFunc:               grafanaregistry.KeyRootFunc(resourceInfo.GroupResource()),
		KeyFunc:                   grafanaregistry.NamespaceKeyFunc(resourceInfo.GroupResource()),
		PredicateFunc:             grafanaregistry.Matcher,
		DefaultQualifiedResource:  resourceInfo.GroupResource(),
		SingularQualifiedResource: resourceInfo.SingularGroupResource(),
		CreateStrategy:            strategy,
		UpdateStrategy:            strategy,
		DeleteStrategy:            strategy,
		TableConvertor:            s.tableConverter,
	}

	options := &generic.StoreOptions{RESTOptions: optsGetter}
	if err := store.CompleteWithOptions(options); err != nil {
		return nil, err
	}
	return store, err
}
