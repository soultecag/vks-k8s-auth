package client

import (
	"fmt"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// ConvertRESTConfigToKubeconfig converts a *rest.Config into a valid kubeconfig YAML string
func ConvertRESTConfigToKubeconfig(clusterName, userName, contextName string, config *rest.Config) (string, error) {

	// 1. Populate the clientcmdapi.Config structure
	kubeconfigObj := clientcmdapi.Config{
		Kind:       "Config",
		APIVersion: "v1",
		Clusters: map[string]*clientcmdapi.Cluster{
			clusterName: {
				Server:                   config.Host,
				CertificateAuthorityData: config.TLSClientConfig.CAData,
				InsecureSkipTLSVerify:    config.TLSClientConfig.Insecure,
			},
		},
		AuthInfos: map[string]*clientcmdapi.AuthInfo{
			userName: {
				Token:    config.BearerToken,
				Username: config.Username,
			},
		},
		Contexts: map[string]*clientcmdapi.Context{
			contextName: {
				Cluster:  clusterName,
				AuthInfo: userName,
			},
		},
		CurrentContext: contextName,
	}

	// 2. Marshal the object to a YAML string
	yamlBytes, err := clientcmd.Write(kubeconfigObj)
	if err != nil {
		return "", fmt.Errorf("failed to serialize kubeconfig: %w", err)
	}

	return string(yamlBytes), nil
}
