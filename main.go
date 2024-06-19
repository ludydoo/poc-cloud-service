package main

import (
	"fmt"
	"net/http"
	"poc-cloud-service/log"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		l := log.FromContext(r.Context())
		l.Info("Hello, World!")
		fmt.Fprintf(w, "Hello, World ")
	})
	http.ListenAndServe(":8080", nil)
}
