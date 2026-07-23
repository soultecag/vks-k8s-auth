package k8s_utils

import (
	"fmt"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// ConvertRESTConfigToKubeconfig converts a *rest.Config into a valid kubeconfig YAML string
func ConvertRESTConfigToKubeconfig(clusterName, userName, contextName string, config *rest.Config) (string, error) {
	// Create clientcmdapi.Config object from the rest.Config
	kubeconfigObj := clientcmdapi.Config{
		Kind:       "Config",
		APIVersion: "v1",
		Clusters: map[string]*clientcmdapi.Cluster{
			clusterName: {
				Server:                   config.Host,
				CertificateAuthorityData: config.CAData,
				InsecureSkipTLSVerify:    config.Insecure,
			},
		},
		AuthInfos: map[string]*clientcmdapi.AuthInfo{
			userName: {
				ClientCertificateData: config.CertData,
				ClientKeyData:         config.KeyData,
				Token:                 config.BearerToken,
				TokenFile:             config.BearerTokenFile,
				Username:              config.Username,
				Password:              config.Password,
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

	// Marshal the object to a YAML string using clientcmd.Write
	yamlBytes, err := clientcmd.Write(kubeconfigObj)
	if err != nil {
		return "", fmt.Errorf("failed to serialize kubeconfig: %w", err)
	}

	return string(yamlBytes), nil
}
