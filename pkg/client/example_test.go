package client_test

import (
	"context"

	"github.com/soultecag/vks-k8s-auth/pkg/client"
	corev1 "k8s.io/api/core/v1"
)

func ExampleNewVksSupervisorAuthClient() {
	cfg := client.VksAuthConfig{
		Endpoint:              "https://supervisor.example.local",
		Username:              "administrator@vsphere.local",
		Password:              "your-password",
		TlsInsecureSkipVerify: false,
	}

	// This example is documentation-only. Replace the placeholder values above
	// with real environment-specific credentials before running it.
	if false {
		vksClient, err := client.NewVksSupervisorAuthClient(cfg)
		if err != nil {
			panic(err)
		}

		var namespaces corev1.NamespaceList
		if err := vksClient.List(context.Background(), &namespaces); err != nil {
			panic(err)
		}
	}
}

func ExampleNewVksGuestClusterAuthClient() {
	cfg := client.VksAuthConfig{
		Endpoint:              "https://supervisor.example.local",
		Username:              "administrator@vsphere.local",
		Password:              "your-password",
		GuestClusterName:      "my-guest-cluster",
		GuestClusterNamespace: "my-namespace",
		TlsInsecureSkipVerify: false,
	}

	// This example is documentation-only. Replace the placeholder values above
	// with real environment-specific credentials before running it.
	if false {
		guestClient, err := client.NewVksGuestClusterAuthClient(cfg)
		if err != nil {
			panic(err)
		}

		var namespaces corev1.NamespaceList
		if err := guestClient.List(context.Background(), &namespaces); err != nil {
			panic(err)
		}

		kubeconfig, err := guestClient.GenerateKubeconfig("guest-cluster", "guest-cluster-context")
		if err != nil {
			panic(err)
		}

		_ = kubeconfig
	}
}
