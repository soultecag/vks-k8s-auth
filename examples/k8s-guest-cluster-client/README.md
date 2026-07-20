# k8s-guest-cluster-client example

Minimal example showing how to:

1. authenticate to Supervisor,
2. build a Kubernetes client for a Tanzu guest cluster,
3. list Kubernetes namespaces in the guest cluster,
4. generate a kubeconfig for the guest cluster.

For library details, see the main README: [../../README.md](../../README.md).

## Required environment variables

- `SUPERVISOR_ENDPOINT` (example: `https://10.2.2.5`)
- `VSPHERE_USERNAME`
- `VSPHERE_PASSWORD`
- `TARGET_CLUSTER` (name of the guest cluster)
- `TARGET_CLUSTER_NAMESPACE` (namespace of the guest cluster)

## Optional environment variables

- `VSPHERE_PORT` (example: `443`)

## Run

```bash
export SUPERVISOR_ENDPOINT=https://10.2.2.5
export VSPHERE_USERNAME=myuser@example.ch
export VSPHERE_PASSWORD=myPass24
export TARGET_CLUSTER=my-guest-cluster
export TARGET_CLUSTER_NAMESPACE=my-namespace
# export VSPHERE_PORT=443

go run .
```

The program prints the namespace names in the guest cluster and prints the generated kubeconfig.
