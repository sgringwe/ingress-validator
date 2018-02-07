package main

import (
	"fmt"
	"os"
	"crypto/tls"
	"crypto/x509"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	slack "github.com/ashwanthkumar/slack-go-webhook"
)

const MINIMUM_DAYS = 45

type HostResult struct {
	Host string
	Certs []x509.Certificate
}

func main() {
	fmt.Println("Starting ingress validation")
	client := newClient()

	// Iterate through every ingress and list out each
	ingresses, err := client.Extensions().Ingresses(metav1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	
	var results []HostResult

	// https://github.com/kubernetes/kubernetes/blob/master/staging/src/k8s.io/api/extensions/v1beta1/types.go#L568
	checkedHosts := make(map[string]struct{})
	for _, ingress := range ingresses.Items {
		// https://github.com/kubernetes/kubernetes/blob/master/staging/src/k8s.io/api/extensions/v1beta1/types.go#L601
		for _, rule := range ingress.Spec.Rules {
			if _, checked := checkedHosts[rule.Host]; checked {
				continue
			}

			// https://github.com/kubernetes/kubernetes/blob/master/staging/src/k8s.io/api/extensions/v1beta1/types.go#L652
			results = append(results, checkHost(rule.Host))
			checkedHosts[rule.Host] = struct{}{}
		}
	}

	processResults(results)

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

func checkHost(host string) (result HostResult) {
	result = HostResult{
		Host:  host,
		Certs: []x509.Certificate{},
	}

	conn, err := tls.Dial("tcp", host + ":443", nil)
	
	if err != nil {
		fmt.Printf("Error checking host %s: %s\n", host, err)
		return
	}
	defer conn.Close()
	
	checkedCerts := make(map[string]struct{})
	for _, chain := range conn.ConnectionState().VerifiedChains {
		for _, cert := range chain {
			if _, checked := checkedCerts[string(cert.Signature)]; checked {
				continue
			}

			checkedCerts[string(cert.Signature)] = struct{}{}
			result.Certs = append(result.Certs, *cert)
		}
	}

	return
}

// Produce a json report with all of the verification info and send to configured location or log
func processResults(results []HostResult) {
	for _, result := range results {
		processResult(result)
	}
}

// Each host has many certs in its chain, so we need to go through each one
// TODO: Send the data as a metric somewhere like datadog? Store raw JSON somewhere? Webhooks?
// For each ingress, verify that it
// - has TLS enabled
// - expires > 45 days
// - only supports tls 1.1, 1.2
func processResult(result HostResult) {
	timeNow := time.Now()
	var minimumDays int64 = 9999

	for _, cert := range result.Certs {
		days := int64(cert.NotAfter.Sub(timeNow).Hours()/24)

		if days < minimumDays {
			minimumDays = days
		}
	}

	fmt.Printf("Host %s certificate expires in %d days\n", result.Host, minimumDays)

	if minimumDays < MINIMUM_DAYS {
		message := fmt.Sprintf("Soon expiring certificate found for host %s! Expires in %d days.", result.Host, minimumDays)
		fmt.Println(message)
		sendSlackMessage(result, message)
	}
}

func sendSlackMessage(result HostResult, message string) {
	slackWebhook := os.Getenv("SLACK_WEBHOOK")

	if slackWebhook == "" {
		fmt.Printf("Unable to send slack webhook for soon expiring certificate for host %s, webhook is missing\n", result.Host)
		return
	}

	payload := slack.Payload {
		Text: message,
		Username: "kubernetes-robot",
		Channel: "#dev-ops",
		IconEmoji: ":lock:",
	}

	err := slack.Send(slackWebhook, "", payload)
	if len(err) > 0 {
		fmt.Printf("Unable to send slack webhook for expiring certificate for host %s: %s\n", result.Host, err)
		return
	}

	return
}