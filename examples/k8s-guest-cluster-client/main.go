package main

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"

	vks_client "github.com/soultecag/vks-k8s-auth/pkg/client"
)

func main() {
	// Example usage of GetK8sClientForGuestCluster
	//
	targetCluster := "my-guest-cluster"
	targetClusterNamespace := "my-namespace"
	supervisorEndpoint := "10.1.1.8"
	username := "username@domain.local"
	password := "myPassw0rd!"

	client, err := vks_client.NewVksGuestClusterAuthClient(vks_client.VksAuthConfig{
		GuestClusterName:      targetCluster,
		GuestClusterNamespace: targetClusterNamespace,
		Endpoint:              supervisorEndpoint,
		Username:              username,
		Password:              password,
		TlsInsecureSkipVerify: true,
	})
	if err != nil {
		fmt.Printf("Error creating Kubernetes client for guest cluster: %v\n", err)
		return
	}

	fmt.Printf("Successfully created Kubernetes client for guest cluster: %v\n", targetCluster)
	// fmt.Printf("REST Config: Host: %v\n", cfg.Host)

	nsList := v1.NamespaceList{}
	client.List(context.Background(), &nsList)
	fmt.Printf("Namespaces in guest cluster:\n")
	for _, ns := range nsList.Items {
		fmt.Printf("- %s\n", ns.Name)
	}

}
