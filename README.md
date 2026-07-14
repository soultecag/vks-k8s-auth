# vks-k8s-auth

Simple Go client for logging into a vSphere Supervisor and creating a Kubernetes client or kubeconfig.

## What this library does

1. Logs in to the Supervisor API using username/password.
2. Reads TLS information from the API server.
3. Builds a Kubernetes client (`controller-runtime` client).
4. Can generate a kubeconfig string from the authenticated session.

## Requirements

- Go 1.26+
- Access to your Supervisor endpoint
- Valid SSO credentials

## Install

```bash
go get github.com/soultec/vks-k8s-auth
```

## Quick start

```go
package main

import (
	"context"
	"fmt"

	"github.com/soultec/vks-k8s-auth/pkg/client"
	corev1 "k8s.io/api/core/v1"
)

func main() {
	cfg := client.VksAuthConfig{
		Endpoint: "https://10.0.0.10",
		Username: "administrator@vsphere.local",
		Password: "your-password",
	}

	vksClient, err := client.NewVksK8sAuthClient(cfg)
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

## Configuration

`VksAuthConfig` fields:

- `Endpoint`: Supervisor URL or host.
- `Port`: optional port override.
- `Username`: SSO username.
- `Password`: SSO password.
- `TlsInsecureSkipVerify`: disable TLS verification for test environments.

## Example

A runnable example is in `examples/k8s-client`.

See `examples/k8s-client/README.md` for steps.

## Acknowledgements

This project is heavily inspired by the excellent work of  **[William Arroyo (@warroyo)](https://github.com/warroyo)** and his Supervisor login examples:

- https://github.com/warroyo/supervisor-login-examples

The authentication flow and interaction with the vSphere Supervisor API are based on the concepts demonstrated in that project. This library builds upon those ideas by providing a Go package that:

- exposes a reusable API for applications and operators
- creates a `controller-runtime` Kubernetes client
- generates kubeconfig files programmatically
- is intended to be consumed as a Go module

Many thanks to William Arroyo for publishing the original examples and making them available to the community.
