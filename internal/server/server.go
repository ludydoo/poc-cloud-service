package server

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/rs/xid"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	v1 "poc-cloud-service/gen/api/v1"
	"poc-cloud-service/internal/argocd"
	"poc-cloud-service/internal/constants"
	"poc-cloud-service/internal/convert"
	"poc-cloud-service/internal/store"
	"poc-cloud-service/log"
	"time"
)

type Server struct {
	v1.UnimplementedTenantServiceServer
	client   kubernetes.Interface
	db       *pgx.Conn
	store    *store.Queries
	informer informers.GenericInformer
}

func NewServer(ctx context.Context, client kubernetes.Interface, dynamicClient dynamic.Interface, store *store.Queries) (*Server, error) {
	l := log.FromContext(ctx)
	factory := dynamicinformer.NewDynamicSharedInformerFactory(dynamicClient, time.Hour)
	informer := factory.ForResource(constants.ArgoApplicationsGVR)

	l.Info("starting informer")
	go func() {
		informer.Informer().Run(ctx.Done())
	}()

	l.Info("waiting for cache sync")
	factory.WaitForCacheSync(ctx.Done())

	l.Info("cache sync done")
	return &Server{
		client:   client,
		store:    store,
		informer: informer,
	}, nil
}

func (s *Server) CreateTenant(ctx context.Context, request *v1.CreateTenantRequest) (*v1.CreateTenantResponse, error) {
	id := xid.New().String()
	valuesJson, err := request.GetSource().GetHelm().GetValues().MarshalJSON()
	if err != nil {
		return nil, err
	}
	created, err := s.store.CreateTenant(ctx, store.CreateTenantParams{
		ID:             id,
		RepoUrl:        request.GetSource().RepoUrl,
		Path:           request.GetSource().GetPath(),
		Values:         valuesJson,
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

	if err := s.withApplication(resp.Tenant); err != nil {
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
		if err := s.withApplication(tenant); err != nil {
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
		ID:             request.Id,
		RepoUrl:        request.GetSource().GetRepoUrl(),
		Path:           request.GetSource().GetPath(),
		Values:         valuesJson,
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

func (s *Server) withApplication(tenant *v1.Tenant) error {
	app, err := s.getApplication(tenant.GetId())
	if err != nil {
		return err
	}
	if app == nil {
		return nil
	}
	tenant.Application = &v1.Application{
		Health: &v1.Health{
			Status:  app.Status.Health.Status,
			Message: app.Status.Health.Message,
		},
	}
	return nil
}

func (s *Server) getApplication(tenantID string) (*argocd.Application, error) {
	obj, err := s.informer.Lister().Get(fmt.Sprintf("%s/%s",
		constants.OpenshiftGitopsNamespace,
		constants.ApplicationNameForTenant(tenantID),
	))
	if err != nil {
		if !errors.IsNotFound(err) {
			return nil, err
		}
		return nil, nil
	}
	unstruct, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return nil, fmt.Errorf("unexpected object type: %T", obj)
	}
	return argocd.FromUnstructured(unstruct)
}
