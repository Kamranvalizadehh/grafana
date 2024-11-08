package server

import (
	"context"

	openfgav1 "github.com/openfga/api/proto/openfga/v1"

	"github.com/grafana/grafana/pkg/services/authz/zanzana/common"
	authzextv1 "github.com/grafana/grafana/pkg/services/authz/zanzana/proto/v1"
)

func (s *Server) Write(ctx context.Context, req *authzextv1.WriteRequest) (*authzextv1.WriteResponse, error) {
	ctx, span := tracer.Start(ctx, "authzServer.Write")
	defer span.End()

	storeInf, err := s.getNamespaceStore(ctx, req.Namespace)
	if err != nil {
		return nil, err
	}
	if storeInf.AuthorizationModelId == "" {
		return nil, errAuthorizationModelNotInitialized
	}

	writeTuples := make([]*openfgav1.TupleKey, 0)
	for _, t := range req.GetWrites().GetTupleKeys() {
		writeTuples = append(writeTuples, common.ToOpenFGATupleKey(t))
	}

	deleteTuples := make([]*openfgav1.TupleKeyWithoutCondition, 0)
	for _, t := range req.GetDeletes().GetTupleKeys() {
		deleteTuples = append(deleteTuples, common.ToOpenFGATupleKeyWithoutCondition(t))
	}

	writeReq := &openfgav1.WriteRequest{
		StoreId:              storeInf.Id,
		AuthorizationModelId: storeInf.AuthorizationModelId,
	}
	if len(writeTuples) > 0 {
		writeReq.Writes = &openfgav1.WriteRequestWrites{
			TupleKeys: writeTuples,
		}
	}
	if len(deleteTuples) > 0 {
		writeReq.Deletes = &openfgav1.WriteRequestDeletes{
			TupleKeys: deleteTuples,
		}
	}

	_, err = s.openfga.Write(ctx, writeReq)
	if err != nil {
		return nil, err
	}

	return &authzextv1.WriteResponse{}, nil
}
