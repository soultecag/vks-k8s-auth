package main

import (
	"context"
	"os"
	"strconv"

	"github.com/soultecag/vks-k8s-auth/pkg/client"
	v1 "k8s.io/api/core/v1"
)

var (
	port     int
	endpoint string
	username string
	password string
)

func init() {

	//Read the VKS API server endpoint, username, and password from environment variables (SUPERVISOR_ENDPOINT, VSPHERE_USERNAME and VSPHERE_PASSWORD )

	endpoint = os.Getenv("SUPERVISOR_ENDPOINT")
	username = os.Getenv("VSPHERE_USERNAME")
	password = os.Getenv("VSPHERE_PASSWORD")
	portString := os.Getenv("VSPHERE_PORT")

	if endpoint == "" || username == "" || password == "" {
		panic("SUPERVISOR_ENDPOINT, VSPHERE_USERNAME and VSPHERE_PASSWORD environment variables must be set")
	}

	if portString != "" {
		parsedPort, err := strconv.Atoi(portString)
		if err == nil {
			port = parsedPort
		}
	}

}

func main() {
	// Create a new VksK8sAuthClient with the configuration read from environment variables.
	cfg := client.VksAuthConfig{
		TlsInsecureSkipVerify: false, // Set to true for testing purposes. In production, set to false and provide a valid CA certificate.
		Endpoint:              endpoint,
		Username:              username,
		Password:              password,
		Port:                  port, // Use the port read from environment variable or default to 0 if not set.
	}

	vksClient, err := client.NewVksSupervisorAuthClient(cfg)
	if err != nil {
		panic(err)
	}

	// Use the vksClient to interact with the Kubernetes API server.
	list := v1.NamespaceList{}

	err = vksClient.Client.List(context.Background(), &list)
	if err != nil {
		panic(err)
	}

	// Print the names of the namespaces retrieved from the Kubernetes API server.
	for _, ns := range list.Items {
		println(ns.Name)
	}

	config, err := vksClient.GenerateKubeconfig("cluster", "context")
	if err != nil {
		panic(err)
	}

	println("Generated kubeconfig:")
	println(config)

	// write the kubeconfig to a file
	err = os.WriteFile("kubeconfig.yaml", []byte(config), 0644)
	if err != nil {
		panic(err)
	}
	println("Kubeconfig written to kubeconfig.yaml")

}
