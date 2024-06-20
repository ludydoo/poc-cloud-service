package reconciler

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"os"
	"poc-cloud-service/log"
	"reflect"
	"time"
)

var applicationsGVR = schema.GroupVersionResource{
	Group:    "argoproj.io",
	Version:  "v1alpha1",
	Resource: "applications",
}

var applicationGVK = schema.GroupVersionKind{
	Group:   "argoproj.io",
	Version: "v1alpha1",
	Kind:    "Application",
}

const (
	namespacePrefix          = "acs-"
	isTenantLabel            = "is-tenant"
	tenantLabel              = "tenant"
	argoCdManagedBy          = "argocd.argoproj.io/managed-by"
	managedByOpenshiftGitops = "openshift-gitops"
	defaultRepoURL           = "https://github.com/ludydoo/poc-cloud-service-manifests"
	defaultRepoPath          = "tenant-manifests"
	openshiftGitopsNamespace = "openshift-gitops"
)

type Reconciler struct {
	client        kubernetes.Interface
	dynamicClient dynamic.Interface
}

func NewReconciler(client kubernetes.Interface, dynamicClient dynamic.Interface) *Reconciler {
	return &Reconciler{
		client:        client,
		dynamicClient: dynamicClient,
	}
}

func (r *Reconciler) Start(ctx context.Context) {

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	l := log.FromContext(ctx)

	go func() {
		for range ticker.C {
			select {
			case <-ctx.Done():
				return
			default:
				l.Info("Reconciling tenants")
				if err := r.reconcileTenants(ctx); err != nil {
					l.Error("Error reconciling tenants", zap.Error(err))
				}
			}
		}
	}()

	<-ctx.Done()
	l.Info("Shutting down")
}

func (r *Reconciler) reconcileTenants(ctx context.Context) error {

	// Want is the desired state
	want, err := getWant()
	if err != nil {
		return err
	}

	// Create/Update tenants
	for _, tenant := range want {
		if err := r.ensureTenantNamespace(log.WithTenant(ctx, tenant.ID), tenant); err != nil {
			return err
		}
		if err := r.ensureTenantApplication(log.WithTenant(ctx, tenant.ID), tenant); err != nil {
			return err
		}
	}

	// Existing is the current state
	existing, err := r.getExistingTenantIDs(ctx)
	if err != nil {
		return err
	}

	toDelete := map[string]string{}
	for _, tenantID := range existing {
		toDelete[tenantID] = tenantID
	}
	for _, tenant := range want {
		delete(toDelete, tenant.ID)
	}

	// Delete tenants that are not in the desired state
	for _, tenantID := range toDelete {
		if err := r.deleteTenant(log.WithTenant(ctx, tenantID), tenantID); err != nil {
			return err
		}
	}

	return nil
}

type Tenant struct {
	ID     string                 `json:"id"`
	Source map[string]interface{} `json:"source"`
}

// getNamespaceTenant extracts the tenant ID from a namespace object
func getNamespaceTenant(namespace v1.Namespace) (string, error) {
	if namespace.Labels == nil {
		return "", fmt.Errorf("no labels")
	}
	tenant, ok := namespace.Labels[tenantLabel]
	if !ok {
		return "", fmt.Errorf("no tenant label")
	}
	if tenant == "" {
		return "", fmt.Errorf("empty tenant label")
	}
	return tenant, nil
}

// deleteTenant deletes a tenant by deleting the namespace and the application
func (r *Reconciler) deleteTenant(ctx context.Context, tenantID string) error {
	l := log.FromContext(ctx)
	l.Info("Deleting tenant")
	if err := deleteTenantApp(ctx, r.dynamicClient, tenantID); err != nil {
		return err
	}
	if err := deleteTenantNamespace(ctx, r.client, tenantID); err != nil {
		return err
	}
	return nil
}

// deleteTenantApp deletes the tenant application
func deleteTenantApp(ctx context.Context, dynamicClient dynamic.Interface, tenant string) error {

	l := log.FromContext(ctx)

	got, err := dynamicClient.Resource(applicationsGVR).Namespace(openshiftGitopsNamespace).Get(ctx, getTenantNamespaceName(tenant), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			l.Info("Application already deleted", zap.String("name", getTenantNamespaceName(tenant)))
			return nil
		}
		return err
	}

	if got.GetDeletionTimestamp() == nil {
		l.Info("Deleting application", zap.String("name", getTenantNamespaceName(tenant)))
		err = dynamicClient.Resource(applicationsGVR).Namespace(openshiftGitopsNamespace).Delete(ctx, getTenantNamespaceName(tenant), metav1.DeleteOptions{})
		if !errors.IsNotFound(err) {
			return err
		}
		return nil
	}

	// wait for deletion
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timed out waiting for deletion")
		case <-ticker.C:
			got, err = dynamicClient.Resource(applicationsGVR).Namespace(openshiftGitopsNamespace).Get(ctx, getTenantNamespaceName(tenant), metav1.GetOptions{})
			if err != nil {
				if errors.IsNotFound(err) {
					l.Info("Application deleted", zap.String("name", getTenantNamespaceName(tenant)))
					return nil
				}
				return err
			}
		}
	}
}

// deleteTenantNamespace deletes the tenant namespace
func deleteTenantNamespace(ctx context.Context, client kubernetes.Interface, tenantID string) error {
	err := client.CoreV1().Namespaces().Delete(ctx, getTenantNamespaceName(tenantID), metav1.DeleteOptions{})
	if !errors.IsNotFound(err) {
		return err
	}
	return nil
}

// ensureTenantApplication ensures that the tenant application exists
func (r *Reconciler) ensureTenantApplication(ctx context.Context, tenant Tenant) error {
	l := log.FromContext(ctx)
	apps := r.dynamicClient.Resource(applicationsGVR)
	want := makeTenantApplication(tenant)
	got, err := apps.Namespace(openshiftGitopsNamespace).Get(ctx, getTenantNamespaceName(tenant.ID), metav1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
		l.Info("Creating application")
		if _, err := apps.Namespace(openshiftGitopsNamespace).Create(ctx, want, metav1.CreateOptions{}); err != nil {
			return err
		}
		return nil
	}

	gotSpec := got.Object["spec"]
	wantSpec := want.Object["spec"]
	if reflect.DeepEqual(gotSpec, wantSpec) {
		return nil
	}

	// update
	l.Info("Updating application")
	got.Object["spec"] = wantSpec
	if _, err := apps.Namespace(openshiftGitopsNamespace).Update(ctx, got, metav1.UpdateOptions{}); err != nil {
		return err
	}

	return nil

}

// ensureTenantNamespace ensures that a namespace exists with the correct labels
func (r *Reconciler) ensureTenantNamespace(ctx context.Context, tenant Tenant) error {
	l := log.FromContext(ctx)

	wantLabels := map[string]string{
		isTenantLabel:   "true",
		tenantLabel:     tenant.ID,
		argoCdManagedBy: managedByOpenshiftGitops,
	}

	namespaceName := getTenantNamespaceName(tenant.ID)

	got, err := r.client.CoreV1().Namespaces().Get(ctx, namespaceName, metav1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
		l.Info("Creating namespace", zap.String("name", namespaceName))
		if _, err := r.client.CoreV1().Namespaces().Create(ctx, &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:   namespaceName,
				Labels: wantLabels,
			},
		}, metav1.CreateOptions{}); err != nil {
			return err
		}
	}

	shouldUpdate := false

	if got.Labels == nil {
		shouldUpdate = true
		got.Labels = map[string]string{}
	}
	for k, v := range wantLabels {
		if got.Labels[k] != v {
			shouldUpdate = true
			got.Labels[k] = v
		}
	}

	if shouldUpdate {
		l.Info("Updating namespace", zap.String("name", namespaceName))
		if _, err := r.client.CoreV1().Namespaces().Update(ctx, got, metav1.UpdateOptions{}); err != nil {
			return err
		}
	}

	return nil

}

// getTenantNamespaceName returns the namespace name for a tenant
func getTenantNamespaceName(tenantID string) string {
	return fmt.Sprintf("%s%s", namespacePrefix, tenantID)
}

// makeTenantApplication creates an ArgoCD Application object for a tenant
func makeTenantApplication(tenant Tenant) *unstructured.Unstructured {
	u := &unstructured.Unstructured{}
	u.SetNamespace(openshiftGitopsNamespace)
	u.SetName(getTenantNamespaceName(tenant.ID))
	u.SetLabels(map[string]string{
		isTenantLabel: "true",
		tenantLabel:   tenant.ID,
	})
	u.SetGroupVersionKind(applicationGVK)

	source := map[string]interface{}{
		"repoURL": defaultRepoURL,
		"path":    defaultRepoPath,
	}

	if len(tenant.Source) > 0 {
		source = tenant.Source
	}

	if _, ok := source["repoURL"].(string); !ok {
		source["repoURL"] = defaultRepoURL
	}

	if source["helm"] == nil {
		source["helm"] = map[string]interface{}{}
	}
	source["helm"].(map[string]interface{})["releaseName"] = getTenantNamespaceName(tenant.ID)

	u.Object["spec"] = map[string]interface{}{
		"project": "default",
		"source":  source,
		"destination": map[string]interface{}{
			"server":    "https://kubernetes.default.svc",
			"namespace": getTenantNamespaceName(tenant.ID),
		},
		"syncPolicy": map[string]interface{}{
			"automated": map[string]interface{}{
				"prune":    true,
				"selfHeal": true,
			},
		},
	}

	return u
}

// getExistingTenantIDs returns a list of existing tenants ids
func (r *Reconciler) getExistingTenantIDs(ctx context.Context) ([]string, error) {
	namespaces, err := r.client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=true", isTenantLabel),
	})
	if err != nil {
		return nil, err
	}

	var tenants []string
	for _, ns := range namespaces.Items {
		tenantID, err := getNamespaceTenant(ns)
		if err != nil {
			return nil, err
		}
		tenants = append(tenants, tenantID)
	}

	return tenants, nil
}

// getWant reads the desired state from a file
func getWant() ([]Tenant, error) {
	dataFile, err := os.ReadFile("/data/tenants")
	if err != nil {
		fmt.Printf("Error: %v", err)
		return nil, err
	}

	var tenants []Tenant
	err = yaml.Unmarshal(dataFile, &tenants)
	if err != nil {
		fmt.Printf("Error: %v", err)
		return nil, err
	}

	return tenants, nil
}
