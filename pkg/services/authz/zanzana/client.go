package zanzana

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	authzlib "github.com/grafana/authlib/authz"
	openfgav1 "github.com/openfga/api/proto/openfga/v1"

	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/services/authz/zanzana/client"
	authzextv1 "github.com/grafana/grafana/pkg/services/authz/zanzana/proto/v1"
	"github.com/grafana/grafana/pkg/setting"
)

// Client is a wrapper around [openfgav1.OpenFGAServiceClient]
type Client interface {
	Check(ctx context.Context, in *openfgav1.CheckRequest) (*openfgav1.CheckResponse, error)
	Read(ctx context.Context, in *openfgav1.ReadRequest) (*openfgav1.ReadResponse, error)
	ListObjects(ctx context.Context, in *openfgav1.ListObjectsRequest) (*openfgav1.ListObjectsResponse, error)
	Write(ctx context.Context, in *openfgav1.WriteRequest) error
}

type ExtensionClient interface {
	authzlib.AccessChecker
	Write(ctx context.Context, req *authzextv1.WriteRequest) (*authzextv1.WriteResponse, error)
}

func NewClient(ctx context.Context, cc grpc.ClientConnInterface, cfg *setting.Cfg) (*client.Client, error) {
	return client.New(
		ctx,
		cc,
		client.WithTenantID(fmt.Sprintf("stack-%s", cfg.StackID)),
		client.WithLogger(log.New("zanzana-client")),
	)
}

func NewNoopClient() *client.NoopClient {
	return client.NewNoop()
}
