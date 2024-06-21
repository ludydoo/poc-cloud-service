package argocd

import (
	"encoding/json"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Application struct {
	TypeMeta   metav1.TypeMeta   `json:",inline"`
	ObjectMeta metav1.ObjectMeta `json:"metadata"`
	Spec       ApplicationSpec   `json:"spec"`
	Status     Status            `json:"status"`
}

type ApplicationSpec struct {}

type HealthStatus struct {
	Status string `json:"status"`
	Message string `json:"message"`
}

type Status struct {
	Health HealthStatus `json:"health"`
}

func FromUnstructured(obj *unstructured.Unstructured) (*Application, error) {
	var app Application
	jsonBytes, err := obj.MarshalJSON()
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(jsonBytes, &app); err != nil {
		return nil, err
	}
	return &app, nil
}
