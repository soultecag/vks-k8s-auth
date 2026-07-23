// Package client provides authentication helpers for connecting to a vSphere
// Supervisor or Tanzu guest cluster and constructing Kubernetes clients from
// the resulting session.
//
// The package logs in to the Supervisor API, captures the API server's TLS
// configuration, and exposes helpers for:
//
//   - creating a controller-runtime client for the Supervisor cluster
//   - creating a controller-runtime client for a guest cluster
//   - generating kubeconfig content from an authenticated session
//   - inspecting and refreshing the session JWT token
//
// Typical usage starts with a VksAuthConfig passed to either
// NewVksSupervisorAuthClient or NewVksGuestClusterAuthClient.
package client
