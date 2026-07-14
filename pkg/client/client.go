package client

import (
	"fmt"

	"k8s.io/client-go/rest"
	k8sapiClient "sigs.k8s.io/controller-runtime/pkg/client"
)

type VksK8sAuthClient struct {
	k8sapiClient.Client
	cfg VksAuthConfig
	// JWT Token received from the VKS API server after successful authentication.
	token string
	// TLSClientConfig is the TLS configuration for the VKS API server.
	tlsConfig rest.TLSClientConfig
}

type VksAuthConfig struct {
	// TlsInsecureSkipVerify is a flag to skip TLS verification for the VKS API server.
	TlsInsecureSkipVerify bool
	// Endpoint is the URL of the VKS Supervisor API server. E.g. https://10.5.24.5
	Endpoint string
	// Port is the port of the VKS Supervisor API server. If not specified, the default port will be used.
	Port int
	// Username is the username to use for authentication with the VKS API server.
	Username string
	// Password is the password to use for authentication with the VKS API server.
	Password string
}

func NewVksK8sAuthClient(config VksAuthConfig) (*VksK8sAuthClient, error) {
	// Validate the supervisor endpoint and port and format it correctly
	host, err := getSupervisorHost(config.Endpoint, config.Port)
	if err != nil {
		return nil, fmt.Errorf("get supervisor host: %w", err)
	}
	config.Endpoint = host

	client := &VksK8sAuthClient{
		cfg: config,
	}

	// Perform login to get the token and initialize the Kubernetes client.
	if err := client.Login(); err != nil {
		return nil, err
	}

	client.tlsConfig, err = client.buildTLSConfig()
	if err != nil {
		return nil, fmt.Errorf("build TLS config failed: %w", err)
	}

	kubeConfig, err := client.buildSupervisorKubeconfig()
	if err != nil {
		return nil, fmt.Errorf("create kubeconfig failed: %w", err)
	}

	kubeClient, err := k8sapiClient.New(kubeConfig, k8sapiClient.Options{})
	if err != nil {
		return nil, fmt.Errorf("create kubernetes client failed: %w", err)
	}

	client.Client = kubeClient

	return client, nil
}

// Login performs the login to the VKS API server and initializes the Kubernetes client.
func (c *VksK8sAuthClient) Login() error {
	// Calls the login method to get the token and store it in c.token.
	token, raw, err := c.login()
	if err != nil {
		return fmt.Errorf("login failed: %w, response: %s", err, raw)
	}
	c.token = token

	return nil
}

func (c *VksK8sAuthClient) GenerateKubeconfig(clusterName, contextName string) (kubeConfig string, err error) {

	kubeConfig, err = ConvertRESTConfigToKubeconfig(clusterName, c.cfg.Username, contextName, &rest.Config{
		Host:            c.cfg.Endpoint,
		BearerToken:     c.token,
		TLSClientConfig: c.tlsConfig,
	})

	return
}
