package k8s_utils_test

import (
	"testing"

	"github.com/soultecag/vks-k8s-auth/pkg/k8s_utils"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func TestConvertRESTConfigToKubeconfig(t *testing.T) {
	kubeConfigYAML, err := k8s_utils.ConvertRESTConfigToKubeconfig("test-cluster", "test-user", "test-context", &rest.Config{
		Host:        "https://10.5.24.5:6443",
		BearerToken: "test-token",
		TLSClientConfig: rest.TLSClientConfig{
			CAData: []byte("test-ca-data"),
		},
	})
	if err != nil {
		t.Fatalf("ConvertRESTConfigToKubeconfig() failed: %v", err)
	}

	parsed, err := clientcmd.Load([]byte(kubeConfigYAML))
	if err != nil {
		t.Fatalf("failed to parse generated kubeconfig: %v", err)
	}

	if parsed.CurrentContext != "test-context" {
		t.Errorf("CurrentContext = %q, want %q", parsed.CurrentContext, "test-context")
	}

	cluster, ok := parsed.Clusters["test-cluster"]
	if !ok {
		t.Fatal("kubeconfig does not contain expected cluster entry")
	}
	if cluster.Server != "https://10.5.24.5:6443" {
		t.Errorf("Cluster.Server = %q, want %q", cluster.Server, "https://10.5.24.5:6443")
	}
	if string(cluster.CertificateAuthorityData) != "test-ca-data" {
		t.Errorf("Cluster.CertificateAuthorityData = %q, want %q", cluster.CertificateAuthorityData, "test-ca-data")
	}

	authInfo, ok := parsed.AuthInfos["test-user"]
	if !ok {
		t.Fatal("kubeconfig does not contain expected auth info entry")
	}
	if authInfo.Token != "test-token" {
		t.Errorf("AuthInfo.Token = %q, want %q", authInfo.Token, "test-token")
	}

	context, ok := parsed.Contexts["test-context"]
	if !ok {
		t.Fatal("kubeconfig does not contain expected context entry")
	}
	if context.Cluster != "test-cluster" || context.AuthInfo != "test-user" {
		t.Errorf("Context = %+v, want Cluster=test-cluster AuthInfo=test-user", context)
	}
}

func TestConvertRESTConfigToKubeconfig_NoCredentialLeakageOnEmptyConfig(t *testing.T) {
	// Guards against accidentally serializing zero-value secrets (e.g. an
	// empty token/password ending up as a literal empty string is fine, but
	// this pins the current, safe behavior against regressions).
	kubeConfigYAML, err := k8s_utils.ConvertRESTConfigToKubeconfig("c", "u", "ctx", &rest.Config{Host: "https://example.com"})
	if err != nil {
		t.Fatalf("ConvertRESTConfigToKubeconfig() failed: %v", err)
	}

	parsed, err := clientcmd.Load([]byte(kubeConfigYAML))
	if err != nil {
		t.Fatalf("failed to parse generated kubeconfig: %v", err)
	}
	if parsed.AuthInfos["u"].Token != "" {
		t.Error("expected empty token to remain empty in generated kubeconfig")
	}
}
