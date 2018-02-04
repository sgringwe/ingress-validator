package main

import (
	"fmt"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {
	fmt.Println("Starting ingress validation")
	client := newClient()

	// Iterate through every ingress and list out each
	ingresses, err := client.Extensions().Ingresses(metav1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("There are %d ingresses in the cluster\n", len(ingresses.Items))

	// For each ingress, verify that it
	// - has TLS enabled
	// - expires > 45 days
	// - only supports tls 1.1, 1.2
	
	// Produce a json report with all of the verification info and send to configured location or log

	fmt.Println("Finishing ingress validation")
}

func newClient() *kubernetes.Clientset {
	var err error
	var config *rest.Config
	config, err = rest.InClusterConfig()
	check(err)
	client, err := kubernetes.NewForConfig(config)
	check(err)
	return client
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}