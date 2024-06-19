package main

import (
	"context"
	"fmt"
	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"os"
	"time"
)

var applicationsGVR = schema.GroupVersionResource{
	Group:    "argoproj.io",
	Version:  "v1alpha1",
	Resource: "applications",
}

func main() {

	// in-cluster kubeconfig
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	ctx := context.Background()

	go func() {
		for range ticker.C {
			select {
			case <-ctx.Done():
				return
			default:
				if err := reconcileTenants(ctx, client, dynamicClient); err != nil {
					fmt.Printf("Error: %v", err)
				}
			}
		}
	}()

	<-ctx.Done()

}

func reconcileTenants(ctx context.Context, client *kubernetes.Clientset, dynamicClient *dynamic.DynamicClient) error {

	// Want is the desired state
	want, err := getWant()
	if err != nil {
		return err
	}

	// Existing is the current state
	existing, err := getExisting(ctx, client)
	if err != nil {
		return err
	}

	toDelete := map[string]tenant{}
	for _, tenant := range existing {
		toDelete[tenant.ID] = tenant
	}
	for _, tenant := range want {
		delete(toDelete, tenant.ID)
	}

	// Delete tenants that are not in the desired state
	for _, tenant := range toDelete {
		if err := deleteTenant(ctx, client, dynamicClient, tenant); err != nil {
			return err
		}
	}

	// Create/Update tenants
	for _, tenant := range want {
		if err := ensureTenantNamespace(ctx, client, tenant); err != nil {
			return err
		}
		if err := ensureTenantApplication(ctx, dynamicClient, tenant); err != nil {
			return err
		}
	}

	return nil
}

type tenant struct {
	ID string `json:"id"`
}

func ensureTenantNamespace(ctx context.Context, client kubernetes.Interface, tenant tenant) error {
	_, err := client.CoreV1().Namespaces().Get(ctx, getTenantNamespaceName(tenant), metav1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
		if _, err := client.CoreV1().Namespaces().Create(ctx, &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: getTenantNamespaceName(tenant),
				Labels: map[string]string{
					"is-tenant":                     "true",
					"argocd.argoproj.io/managed-by": "openshift-gitops",
				},
			},
		}, metav1.CreateOptions{}); err != nil {
			return err
		}
	}

	return nil

}

func deleteTenant(ctx context.Context, client kubernetes.Interface, dynamicClient dynamic.Interface, tenant tenant) error {
	if err := deleteTenantApp(ctx, dynamicClient, tenant); err != nil {
		return err
	}
	if err := deleteTenantNamespace(ctx, client, tenant); err != nil {
		return err
	}
	return nil
}

func deleteTenantApp(ctx context.Context, dynamicClient dynamic.Interface, tenant tenant) error {
	err := dynamicClient.
		Resource(applicationsGVR).
		Namespace("openshift-gitops").
		Delete(ctx, getTenantNamespaceName(tenant), metav1.DeleteOptions{})
	if !errors.IsNotFound(err) {
		return err
	}
	return nil
}

func deleteTenantNamespace(ctx context.Context, client kubernetes.Interface, tenant tenant) error {
	err := client.CoreV1().Namespaces().Delete(ctx, getTenantNamespaceName(tenant), metav1.DeleteOptions{})
	if !errors.IsNotFound(err) {
		return err
	}
	return nil
}

func ensureTenantApplication(ctx context.Context, client dynamic.Interface, tenant tenant) error {
	apps := client.Resource(applicationsGVR)

	want := makeTenantApplication(tenant)
	_, err := apps.Namespace("openshift-gitops").Get(ctx, getTenantNamespaceName(tenant), metav1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
		if _, err := apps.Namespace("openshift-gitops").Create(ctx, want, metav1.CreateOptions{}); err != nil {
			return err
		}
		return nil
	}

	// update
	if _, err := apps.Namespace("openshift-gitops").Update(ctx, want, metav1.UpdateOptions{}); err != nil {
		return err
	}

	return nil

}

func getTenantNamespaceName(tenant tenant) string {
	return fmt.Sprintf("tenant-%s", tenant.ID)
}

func makeTenantApplication(tenant tenant) *unstructured.Unstructured {
	u := &unstructured.Unstructured{}
	u.SetNamespace("openshift-gitops")
	u.SetName(getTenantNamespaceName(tenant))
	u.SetLabels(map[string]string{
		"is-tenant": "true",
	})
	u.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "argoproj.io",
		Version: "v1alpha1",
		Kind:    "Application",
	})

	u.Object["spec"] = map[string]interface{}{
		"project": "default",
		"source": map[string]interface{}{
			"repoURL": "https://github.com/ludydoo/poc-cloud-service-manifests",
			"path":    "tenant-manifests",
		},
		"destination": map[string]interface{}{
			"server":    "https://kubernetes.default.svc",
			"namespace": fmt.Sprintf("tenant-%s", tenant.ID),
		},
		"syncPolicy": map[string]interface{}{
			"automated": map[string]interface{}{
				"prune": true,
			},
		},
	}

	return u
}

func getExisting(ctx context.Context, client kubernetes.Interface) ([]tenant, error) {
	namespaces, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{
		LabelSelector: "is-tenant=true",
	})
	if err != nil {
		return nil, err
	}

	var tenants []tenant
	for _, ns := range namespaces.Items {
		tenants = append(tenants, tenant{
			ID: ns.Name[len("tenant-"):],
		})
	}

	return tenants, nil
}

func getWant() ([]tenant, error) {
	dataFile, err := os.ReadFile("/data/tenants")
	if err != nil {
		fmt.Printf("Error: %v", err)
		return nil, err
	}

	var tenants []tenant
	err = yaml.Unmarshal(dataFile, &tenants)
	if err != nil {
		fmt.Printf("Error: %v", err)
		return nil, err
	}

	return tenants, nil
}
