package main

import (
	"context"
	"fmt"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"time"
)

func main() {

	// in-cluster kubeconfig
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	apps := dynamicClient.Resource(v1alpha1.Resource("applications").WithVersion("v1alpha1"))

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	ctx := context.Background()

	go func() {
		for range ticker.C {
			select {
			case <-ctx.Done():
				return
			default:
				list, err := apps.Namespace("").List(ctx, metav1.ListOptions{})
				if err != nil {
					fmt.Printf("Error: %v", err)
					continue
				}
				for _, item := range list.Items {
					fmt.Printf("Name: %s\n", item.GetName())
				}
			}
		}
	}()

	<-ctx.Done()

}
