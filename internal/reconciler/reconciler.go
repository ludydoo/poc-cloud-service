package reconciler

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	v1 "poc-cloud-service/gen/api/v1"
	"poc-cloud-service/internal/constants"
	"poc-cloud-service/internal/convert"
	"poc-cloud-service/internal/store"
	"poc-cloud-service/log"
	"reflect"
	"sigs.k8s.io/yaml"
	"time"
)

const (
	isTenantLabel            = "is-tenant"
	tenantLabel              = "tenant"
	argoCdManagedBy          = "argocd.argoproj.io/managed-by"
	managedByOpenshiftGitops = "openshift-gitops"
	defaultRepoURL           = "https://github.com/ludydoo/poc-cloud-service-manifests"
	defaultRepoPath          = "tenant-manifests"
)

type Reconciler struct {
	client        kubernetes.Interface
	dynamicClient dynamic.Interface
	store         *store.Queries
}

func NewReconciler(client kubernetes.Interface, dynamicClient dynamic.Interface, store *store.Queries) *Reconciler {
	return &Reconciler{
		client:        client,
		dynamicClient: dynamicClient,
		store:         store,
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
	want, err := r.getWant(ctx)
	if err != nil {
		return fmt.Errorf("failed to get desired state: %w", err)
	}

	// Create/Update tenants
	for _, tenant := range want {
		tenantCtx := log.WithTenant(ctx, tenant.GetId())
		if err := r.ensureTenantNamespace(tenantCtx, tenant); err != nil {
			return fmt.Errorf("failed to ensure tenant namespace: %w", err)
		}
		if err := r.ensureTenantApplication(tenantCtx, tenant); err != nil {
			return fmt.Errorf("failed to ensure tenant application: %w", err)
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
		delete(toDelete, tenant.GetId())
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
func getNamespaceTenant(namespace corev1.Namespace) (string, error) {
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

	got, err := dynamicClient.Resource(constants.ArgoApplicationsGVR).Namespace(constants.OpenshiftGitopsNamespace).Get(ctx, constants.NamespaceNameForTenant(tenant), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			l.Info("Application already deleted", zap.String("name", constants.NamespaceNameForTenant(tenant)))
			return nil
		}
		return err
	}

	if got.GetDeletionTimestamp() == nil {
		l.Info("Deleting application", zap.String("name", constants.NamespaceNameForTenant(tenant)))
		err = dynamicClient.Resource(constants.ArgoApplicationsGVR).Namespace(constants.OpenshiftGitopsNamespace).Delete(ctx, constants.NamespaceNameForTenant(tenant), metav1.DeleteOptions{})
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
			got, err = dynamicClient.Resource(constants.ArgoApplicationsGVR).Namespace(constants.OpenshiftGitopsNamespace).Get(ctx, constants.NamespaceNameForTenant(tenant), metav1.GetOptions{})
			if err != nil {
				if errors.IsNotFound(err) {
					l.Info("Application deleted", zap.String("name", constants.NamespaceNameForTenant(tenant)))
					return nil
				}
				return err
			}
		}
	}
}

// deleteTenantNamespace deletes the tenant namespace
func deleteTenantNamespace(ctx context.Context, client kubernetes.Interface, tenantID string) error {
	err := client.CoreV1().Namespaces().Delete(ctx, constants.NamespaceNameForTenant(tenantID), metav1.DeleteOptions{})
	if !errors.IsNotFound(err) {
		return err
	}
	return nil
}

// ensureTenantApplication ensures that the tenant application exists
func (r *Reconciler) ensureTenantApplication(ctx context.Context, tenant *v1.Tenant) error {
	l := log.FromContext(ctx)
	apps := r.dynamicClient.Resource(constants.ArgoApplicationsGVR)
	want, err := makeTenantApplication(tenant)
	if err != nil {
		return fmt.Errorf("failed to build desired application: %w", err)
	}
	got, err := apps.Namespace(constants.OpenshiftGitopsNamespace).Get(ctx, want.GetName(), metav1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("failed to get application: %w", err)
		}
		l.Info("Creating application")
		if _, err := apps.Namespace(constants.OpenshiftGitopsNamespace).Create(ctx, want, metav1.CreateOptions{}); err != nil {
			return fmt.Errorf("failed to create application: %w", err)
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
	if _, err := apps.Namespace(constants.OpenshiftGitopsNamespace).Update(ctx, got, metav1.UpdateOptions{}); err != nil {
		return fmt.Errorf("failed to update application: %w", err)
	}

	return nil

}

// ensureTenantNamespace ensures that a namespace exists with the correct labels
func (r *Reconciler) ensureTenantNamespace(ctx context.Context, tenant *v1.Tenant) error {
	l := log.FromContext(ctx)

	wantLabels := map[string]string{
		isTenantLabel:   "true",
		tenantLabel:     tenant.GetId(),
		argoCdManagedBy: managedByOpenshiftGitops,
	}

	namespaceName := constants.NamespaceNameForTenant(tenant.GetId())

	got, err := r.client.CoreV1().Namespaces().Get(ctx, namespaceName, metav1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
		l.Info("Creating namespace", zap.String("name", namespaceName))
		if _, err := r.client.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:   namespaceName,
				Labels: wantLabels,
			},
		}, metav1.CreateOptions{}); err != nil {
			return fmt.Errorf("failed to create namespace: %w", err)
		}
		return nil
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
			return fmt.Errorf("failed to update namespace: %w", err)
		}
	}

	return nil

}

// makeTenantApplication creates an ArgoCD Application object for a tenant
func makeTenantApplication(tenant *v1.Tenant) (*unstructured.Unstructured, error) {
	u := &unstructured.Unstructured{}
	u.SetNamespace(constants.OpenshiftGitopsNamespace)
	u.SetName(constants.ApplicationNameForTenant(tenant.GetId()))
	u.SetLabels(map[string]string{
		isTenantLabel: "true",
		tenantLabel:   tenant.GetId(),
	})
	u.SetGroupVersionKind(constants.ArgoApplicationGVK)

	source := map[string]interface{}{
		"repoURL": defaultRepoURL,
		"path":    defaultRepoPath,
	}

	if repoURL := tenant.GetSource().GetRepoUrl(); len(repoURL) > 0 {
		source["repoURL"] = repoURL
	}
	if path := tenant.GetSource().GetPath(); len(path) > 0 {
		source["path"] = path
	}
	if targetRevision := tenant.GetSource().GetTargetRevision(); len(targetRevision) > 0 {
		source["targetRevision"] = targetRevision
	}

	helm := map[string]interface{}{
		"releaseName": constants.NamespaceNameForTenant(tenant.GetId()),
	}

	if values := tenant.GetSource().GetHelm().GetValues().AsMap(); len(values) > 0 {
		yamlBytes, err := yaml.Marshal(values)
		if err != nil {
			return nil, err
		}
		helm["values"] = string(yamlBytes)
	}

	source["helm"] = helm

	u.Object["spec"] = map[string]interface{}{
		"project": "default",
		"source":  source,
		"destination": map[string]interface{}{
			"server":    "https://kubernetes.default.svc",
			"namespace": constants.NamespaceNameForTenant(tenant.GetId()),
		},
		"syncPolicy": map[string]interface{}{
			"automated": map[string]interface{}{
				"prune":    true,
				"selfHeal": true,
			},
		},
	}

	return u, nil
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
func (r *Reconciler) getWant(ctx context.Context) ([]*v1.Tenant, error) {

	storedTenants, err := r.store.ListTenants(ctx)
	if err != nil {
		return nil, err
	}

	tenants, err := convert.TenantsFromStore(storedTenants)
	if err != nil {
		return nil, err
	}

	return tenants, nil
}
