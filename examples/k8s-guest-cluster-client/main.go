package main

import (
	"context"
	"fmt"
	"os"

	v1 "k8s.io/api/core/v1"

	"strconv"

	vks_client "github.com/soultecag/vks-k8s-auth/pkg/client"
)

var (
	port                   int
	endpoint               string
	username               string
	password               string
	targetCluster          string
	targetClusterNamespace string
)

func init() {

	//Read the VKS API server endpoint, username, and password from environment variables (SUPERVISOR_ENDPOINT, VSPHERE_USERNAME and VSPHERE_PASSWORD )

	endpoint = os.Getenv("SUPERVISOR_ENDPOINT")
	username = os.Getenv("VSPHERE_USERNAME")
	password = os.Getenv("VSPHERE_PASSWORD")
	portString := os.Getenv("VSPHERE_PORT")
	targetCluster = os.Getenv("TARGET_CLUSTER")
	targetClusterNamespace = os.Getenv("TARGET_CLUSTER_NAMESPACE")

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

	// Example usage of GetK8sClientForGuestCluster

	client, err := vks_client.NewVksGuestClusterAuthClient(vks_client.VksAuthConfig{
		GuestClusterName:      targetCluster,
		GuestClusterNamespace: targetClusterNamespace,
		Endpoint:              endpoint,
		Port:                  port,
		Username:              username,
		Password:              password,
		TlsInsecureSkipVerify: false,
	})
	if err != nil {
		fmt.Printf("Error creating Kubernetes client for guest cluster: %v\n", err)
		return
	}

	fmt.Printf("Successfully created Kubernetes client for guest cluster: %v\n", targetCluster)
	// fmt.Printf("REST Config: Host: %v\n", cfg.Host)

	nsList := v1.NamespaceList{}
	err = client.List(context.Background(), &nsList)
	if err != nil {
		fmt.Printf("Error listing namespaces in guest cluster: %v\n", err)
		return
	}
	fmt.Printf("Namespaces in guest cluster:\n")
	for _, ns := range nsList.Items {
		fmt.Printf("- %s\n", ns.Name)
	}

	config, err := client.GenerateKubeconfig("cluster", "context")
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
