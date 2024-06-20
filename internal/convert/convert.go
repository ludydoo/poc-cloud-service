package convert

import (
	"encoding/json"
	"fmt"
	"google.golang.org/protobuf/types/known/structpb"
	"poc-cloud-service/gen/api/v1"
	"poc-cloud-service/internal/store"
)

func TenantFromStore(tenant store.Tenant) (*v1.Tenant, error) {
	values := map[string]interface{}{}
	if len(tenant.Values) > 0 {
		if err := json.Unmarshal(tenant.Values, &values); err != nil {
			return nil, err
		}
	}
	helmValues, err := structpb.NewValue(values)
	if err != nil {
		return nil, err
	}
	if helmValues.GetStructValue() == nil {
		return nil, fmt.Errorf("bad helm value format")
	}
	source := &v1.Source{
		RepoUrl: tenant.RepoUrl,
		Path:    tenant.Path,
		Helm: &v1.Helm{
			Values: helmValues.GetStructValue(),
		},
	}
	return &v1.Tenant{
		Id:     tenant.ID,
		Source: source,
	}, nil
}

func TenantsFromStore(tenants []store.Tenant) ([]*v1.Tenant, error) {
	ret := make([]*v1.Tenant, len(tenants))
	for i, tenant := range tenants {
		t, err := TenantFromStore(tenant)
		if err != nil {
			return nil, err
		}
		ret[i] = t
	}
	return ret, nil
}
