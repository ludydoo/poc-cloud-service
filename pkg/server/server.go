package server

import (
	"context"
	"github.com/rs/xid"
	"k8s.io/client-go/kubernetes"
	v1 "poc-cloud-service/gen/api/v1"
)

type Server struct {
	v1.UnimplementedTenantServiceServer
	client kubernetes.Interface
}

func NewServer(client kubernetes.Interface) *Server {
	return &Server{client: client}
}

func (s *Server) CreateTenant(ctx context.Context, request *v1.CreateTenantRequest) (*v1.CreateTenantResponse, error) {
	tenantID := xid.New()
	tenant := v1.Tenant{
		Id:     tenantID.String(),
		Source: request.Source,
	}
	return nil, nil
}

func (s *Server) GetTenant(ctx context.Context, request *v1.GetTenantRequest) (*v1.GetTenantResponse, error) {
	return nil, nil
}

func (s *Server) ListTenants(ctx context.Context, request *v1.ListTenantsRequest) (*v1.ListTenantsResponse, error) {
	return nil, nil
}

func (s *Server) UpdateTenant(ctx context.Context, request *v1.UpdateTenantRequest) (*v1.UpdateTenantResponse, error) {
	return nil, nil
}

func (s *Server) DeleteTenant(ctx context.Context, request *v1.DeleteTenantRequest) (*v1.DeleteTenantResponse, error) {
	return nil, nil
}
