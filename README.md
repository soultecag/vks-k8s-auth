[![Go Reference](https://pkg.go.dev/badge/github.com/soultecag/vks-k8s-auth.svg)](https://pkg.go.dev/github.com/soultecag/vks-k8s-auth)
[![Go Version](https://img.shields.io/github/go-mod/go-version/soultecag/vks-k8s-auth)](https://github.com/soultecag/vks-k8s-auth/blob/main/go.mod)
[![Lint](https://github.com/soultecag/vks-k8s-auth/actions/workflows/lint.yaml/badge.svg)](https://github.com/soultecag/vks-k8s-auth/actions/workflows/lint.yaml)
[![Security Scan](https://github.com/soultecag/vks-k8s-auth/actions/workflows/scan.yaml/badge.svg)](https://github.com/soultecag/vks-k8s-auth/actions/workflows/scan.yaml)
[![License](https://img.shields.io/github/license/soultecag/vks-k8s-auth)](https://github.com/soultecag/vks-k8s-auth/blob/main/LICENSE)
[![Release](https://img.shields.io/github/v/release/soultecag/vks-k8s-auth)](https://github.com/soultecag/vks-k8s-auth/releases)

# vks-k8s-auth

Simple Go client for logging into a vSphere Supervisor and creating a Kubernetes client or kubeconfig, for either the Supervisor cluster itself or one of its Tanzu guest clusters.

## What this library does

1. Logs in to the Supervisor API using username/password.
2. Reads TLS information from the API server.
3. Builds a Kubernetes client (`controller-runtime` client) for the Supervisor, or for a Tanzu guest cluster running on it.
4. Can generate a kubeconfig string from the authenticated session.
5. Handles JWT token expiration and refresh.

## Requirements

- Go 1.26+
- Access to your Supervisor endpoint
- Valid credentials

## Install

```bash
go get github.com/soultecag/vks-k8s-auth
```

## Quick start

```go
package main

import (
	"context"
	"fmt"

	"github.com/soultecag/vks-k8s-auth/pkg/client"
	corev1 "k8s.io/api/core/v1"
)

func main() {
	cfg := client.VksAuthConfig{
		Endpoint: "https://10.0.0.10",
		Username: "administrator@vsphere.local",
		Password: "your-password",
		// Set GuestClusterName/GuestClusterNamespace and use NewVksGuestClusterAuthClient
		// instead to target a Tanzu guest cluster rather than the Supervisor.
	}

	vksClient, err := client.NewVksSupervisorAuthClient(cfg)
	if err != nil {
		panic(err)
	}

	nsList := corev1.NamespaceList{}
	if err := vksClient.List(context.Background(), &nsList); err != nil {
		panic(err)
	}

	fmt.Printf("namespaces: %d\n", len(nsList.Items))
}
```

## Configuration & methods

`VksAuthConfig` fields: `Endpoint` (Supervisor URL/host), `Port` (optional override), `Username`, `Password`, `TlsInsecureSkipVerify`, `Timeout` (login timeout in seconds, defaults to 20), and `GuestClusterName`/`GuestClusterNamespace` (required for `NewVksGuestClusterAuthClient`).

Client methods:

- `NewVksSupervisorAuthClient(cfg)` / `NewVksGuestClusterAuthClient(cfg)`: authenticate and return a client scoped to the Supervisor or a Tanzu guest cluster.
- `GenerateKubeconfig(clusterName, contextName)`: generates a kubeconfig string for the authenticated session.
- `GetToken()`, `TokenValid()`, `TokenExpiry()`, `RefreshToken()`: inspect or refresh the JWT token.
- `ResetHTTPClient()`: closes idle connections and discards the cached HTTP client, so the next login call builds a fresh one (e.g. after changing TLS settings or to force a new connection).

## Examples

Runnable examples are provided in:

- `examples/k8s-client` — authenticate against the Supervisor cluster. See [examples/k8s-client/README.md](examples/k8s-client/README.md).
- `examples/k8s-guest-cluster-client` — authenticate against a Tanzu guest cluster.

## Acknowledgements

This project is heavily inspired by the excellent work of  **[William Arroyo (@warroyo)](https://github.com/warroyo)** and his Supervisor login examples:

- https://github.com/warroyo/supervisor-login-examples

The authentication flow and interaction with the vSphere Supervisor API are based on the concepts demonstrated in that project. This library builds upon those ideas by providing a Go package that:

- exposes a reusable API for applications and operators
- creates a `controller-runtime` Kubernetes client
- generates kubeconfig files programmatically
- is intended to be consumed as a Go module

Many thanks to William Arroyo for publishing the original examples and making them available to the community.
