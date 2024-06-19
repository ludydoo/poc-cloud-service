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
				tenants, err := getTenants()
				if err != nil {
					fmt.Printf("Error: %v", err)
					continue
				}

				for _, tenant := range tenants {
					if err := ensureTenantNamespace(ctx, client, tenant); err != nil {
						fmt.Printf("Error: %v", err)
						return
					}
					if err := ensureTenantApplication(ctx, dynamicClient, tenant); err != nil {
						fmt.Printf("Error: %v", err)
						return
					}
				}
			}
		}
	}()

	<-ctx.Done()

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

func ensureTenantApplication(ctx context.Context, client dynamic.Interface, tenant tenant) error {
	apps := client.Resource(schema.GroupVersionResource{
		Group:    "argoproj.io",
		Version:  "v1alpha1",
		Resource: "applications",
	})

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
			"automated": map[string]interface{}{},
		},
	}

	return u
}

func getTenants() ([]tenant, error) {
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
