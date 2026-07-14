# k8s-client example

Minimal example showing how to:

1. authenticate to Supervisor,
2. list Kubernetes namespaces,
3. generate a kubeconfig file.

For library details, see the main README: [../../README.md](../../README.md).

## Required environment variables

- `SUPERVISOR_ENDPOINT` (example: `https://10.2.2.5`)
- `VSPHERE_USERNAME`
- `VSPHERE_PASSWORD`

## Optional environment variables

- `SUPERVISOR_PORT` (example: `443`)

## Run

```bash
export SUPERVISOR_ENDPOINT=https://10.2.2.5
export VSPHERE_USERNAME=myuser@example.ch
export VSPHERE_PASSWORD=myPass24
# export SUPERVISOR_PORT=443

go run .
```

The program prints namespace names and writes `kubeconfig.yaml` in this folder.