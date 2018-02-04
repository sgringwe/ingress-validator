package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	logger := stdlog.New(os.Stdout, os.Stderr, 0)
	client := newClient()

	// Iterate through every ingress and list out each

	// For each ingress, verify that it
	// - has TLS enabled
	// - expires > 45 days
	// - only supports tls 1.1, 1.2
	
	// Produce a json report with all of the verification info and send to configured location or log
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