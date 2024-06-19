package main

import (
	"fmt"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	// argocd client
	_ "github.com/argoproj/argo-cd/v2/pkg/apiclient"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
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

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		list, err := apps.Namespace("").List(r.Context(), metav1.ListOptions{})
		if err != nil {
			fmt.Fprintf(w, "Error: %v", err)
			return
		}

		for _, item := range list.Items {
			fmt.Fprintf(w, "Name: %s\n", item.GetName())
		}

	})
	http.ListenAndServe(":8080", nil)
}
