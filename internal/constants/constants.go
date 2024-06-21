package constants

import (
	"fmt"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var ArgoApplicationsGVR = schema.GroupVersionResource{
	Group:    "argoproj.io",
	Version:  "v1alpha1",
	Resource: "applications",
}

var ArgoApplicationGVK = schema.GroupVersionKind{
	Group:   "argoproj.io",
	Version: "v1alpha1",
	Kind:    "Application",
}

const TenantNamespacePrefix = "acs-"

func NamespaceNameForTenant(tenantID string) string {
	return fmt.Sprintf("%s%s", TenantNamespacePrefix, tenantID)
}

func ApplicationNameForTenant(tenantID string) string {
	return fmt.Sprintf("%s%s", TenantNamespacePrefix, tenantID)
}

const OpenshiftGitopsNamespace = "openshift-gitops"

