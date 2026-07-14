package main

import (
	"context"
	"os"

	"github.com/soultec/vks-k8s-auth/pkg/client"
	v1 "k8s.io/api/core/v1"
)

func init() {
	//Read the VKS API server endpoint, username, and password from environment variables (SUPERVISOR_ENDPOINT, VSPHERE_USERNAME and VSPHERE_PASSWORD )

	endpoint := os.Getenv("SUPERVISOR_ENDPOINT")
	username := os.Getenv("VSPHERE_USERNAME")
	password := os.Getenv("VSPHERE_PASSWORD")

	if endpoint == "" || username == "" || password == "" {
		panic("SUPERVISOR_ENDPOINT, VSPHERE_USERNAME and VSPHERE_PASSWORD environment variables must be set")
	}

}

func main() {
	// Create a new VksK8sAuthClient with the configuration read from environment variables.
	cfg := client.VksAuthConfig{
		TlsInsecureSkipVerify: false, // Set to true for testing purposes. In production, set to false and provide a valid CA certificate.
		Endpoint:              os.Getenv("SUPERVISOR_ENDPOINT"),
		Username:              os.Getenv("VSPHERE_USERNAME"),
		Password:              os.Getenv("VSPHERE_PASSWORD"),
	}

	vksClient, err := client.NewVksK8sAuthClient(cfg)
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

	config, err := vksClient.GenerateKubeconfig()
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
