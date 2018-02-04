package main

import (
	"fmt"
	"os"
	"crypto/tls"
	"time"

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

	// https://github.com/kubernetes/kubernetes/blob/master/staging/src/k8s.io/api/extensions/v1beta1/types.go#L568
	for _, ingress := range ingresses.Items {
		// https://github.com/kubernetes/kubernetes/blob/master/staging/src/k8s.io/api/extensions/v1beta1/types.go#L601
		for _, rule := range ingress.Spec.Rules {
			// https://github.com/kubernetes/kubernetes/blob/master/staging/src/k8s.io/api/extensions/v1beta1/types.go#L652
			fmt.Printf("Found host %s\n", rule.Host)

			// TODO(sgringwe): How to detect what port the ingress is on? Code comments state that either port
			// 80 or 443 are implied, but not clear how.
			checkHost(rule.Host + ":443")
		}
	}
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

func checkHost(host string) (days int) {
	conn, err := tls.Dial("tcp", host, nil)
	
	if err != nil {
		fmt.Printf("Error checking host %s: %s\n", host, err)
		return
	}
	defer conn.Close()
	
	timeNow := time.Now()
	checkedCerts := make(map[string]struct{})
	for _, chain := range conn.ConnectionState().VerifiedChains {
		for _, cert := range chain {
			if _, checked := checkedCerts[string(cert.Signature)]; checked {
				continue
			}
			checkedCerts[string(cert.Signature)] = struct{}{}

			// Check the expiration compared to 45 days from now
			days := int64(cert.NotAfter.Sub(timeNow).Hours()/24)
			fmt.Printf("Host %s expires in %d days\n", host, days)
			if timeNow.AddDate(0, 0, 45).After(cert.NotAfter) {
				// Send some sort of alert
			}
		}
	}

	return
}