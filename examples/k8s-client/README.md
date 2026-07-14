# k8s-client

Example for usage of the vks-k8s-auth package as k8s-client.

# How To

Export SUPERVISOR_ENDPOINT, VSPHERE_USERNAME and VSPHERE_PASSWORD as env.  
Execute main.go:

```
export SUPERVISOR_ENDPOINT=https://10.2.2.5
export VSPHERE_USERNAME=myuser@example.ch
export VSPHERE_PASSWORD=myPass24
go run main.go
```