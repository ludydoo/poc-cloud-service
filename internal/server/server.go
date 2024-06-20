package server

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/rs/xid"
	"k8s.io/client-go/kubernetes"
	v1 "poc-cloud-service/gen/api/v1"
	"poc-cloud-service/internal/convert"
	"poc-cloud-service/internal/store"
)

type Server struct {
	v1.UnimplementedTenantServiceServer
	client kubernetes.Interface
	db     *pgx.Conn
	store  *store.Queries
}

func NewServer(client kubernetes.Interface, store *store.Queries) *Server {
	return &Server{
		client: client,
		store:  store,
	}
}

func (s *Server) CreateTenant(ctx context.Context, request *v1.CreateTenantRequest) (*v1.CreateTenantResponse, error) {
	id := xid.New().String()
	valuesJson, err := request.GetSource().GetHelm().GetValues().MarshalJSON()
	if err != nil {
		return nil, err
	}
	created, err := s.store.CreateTenant(ctx, store.CreateTenantParams{
		ID:      id,
		RepoUrl: request.GetSource().RepoUrl,
		Path:    request.GetSource().GetPath(),
		Values:  valuesJson,
		TargetRevision: request.GetSource().GetTargetRevision(),
	})
	if err != nil {
		return nil, err
	}
	resp := &v1.CreateTenantResponse{}
	resp.Tenant, err = convert.TenantFromStore(created)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *Server) GetTenant(ctx context.Context, request *v1.GetTenantRequest) (*v1.GetTenantResponse, error) {
	storedTenant, err := s.store.GetTenantByID(ctx, request.GetId())
	if err != nil {
		return nil, err
	}
	resp := &v1.GetTenantResponse{}
	resp.Tenant, err = convert.TenantFromStore(storedTenant)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *Server) ListTenants(ctx context.Context, request *v1.ListTenantsRequest) (*v1.ListTenantsResponse, error) {
	tenants, err := s.store.ListTenants(ctx)
	if err != nil {
		return nil, err
	}

	resp := &v1.ListTenantsResponse{
		Tenants: make([]*v1.Tenant, 0, len(tenants)),
	}

	for _, storedTenant := range tenants {
		tenant, err := convert.TenantFromStore(storedTenant)
		if err != nil {
			return nil, err
		}
		resp.Tenants = append(resp.Tenants, tenant)
	}

	return resp, nil
}

func (s *Server) UpdateTenant(ctx context.Context, request *v1.UpdateTenantRequest) (*v1.UpdateTenantResponse, error) {
	_, err := s.store.GetTenantByID(ctx, request.GetId())
	if err != nil {
		return nil, err
	}
	valuesJson, err := request.GetSource().GetHelm().GetValues().MarshalJSON()
	if err != nil {
		return nil, err
	}
	updated, err := s.store.UpdateTenant(ctx, store.UpdateTenantParams{
		ID:        request.Id,
		RepoUrl:   request.GetSource().GetRepoUrl(),
		Path:      request.GetSource().GetPath(),
		Values:    valuesJson,
		TargetRevision: request.GetSource().GetTargetRevision(),
	})
	if err != nil {
		return nil, err
	}
	resp := &v1.UpdateTenantResponse{}
	resp.Tenant, err = convert.TenantFromStore(updated)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *Server) DeleteTenant(ctx context.Context, request *v1.DeleteTenantRequest) (*v1.DeleteTenantResponse, error) {
	deleted, err := s.store.DeleteTenant(ctx, request.GetId())
	if err != nil {
		return nil, err
	}
	resp := &v1.DeleteTenantResponse{}
	resp.Tenant, err = convert.TenantFromStore(deleted)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
