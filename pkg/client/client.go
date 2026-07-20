package client

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
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

var timeout = 20

type VksAuthConfig struct {
	// TlsInsecureSkipVerify is a flag to skip TLS verification for the VKS API server.
	TlsInsecureSkipVerify bool
	// Endpoint is the URL of the VKS API server. E.g. https://10.5.24.5
	Endpoint string
	// Port is the port of the VKS API server. If not specified, the default port will be used.
	Port int
	// Username is the username to use for authentication with the VKS API server.
	Username string
	// Password is the password to use for authentication with the VKS API server.
	Password string
	// Timeout in Seconds is the timeout for login to the supervisor.
	// If not specified, the default timeout (20) will be used.
	Timeout int

	// GuestClusterName is the name of the guest cluster for which to generate a kube client.
	GuestClusterName string
	// GuestClusterNamespace is the namespace of the guest cluster for which to generate a kube client.
	GuestClusterNamespace string
}

// NewVksSupervisorAuthClient creates a new VksK8sAuthClient with the provided configuration.
// It performs the login to the VKS API server and initializes the Kubernetes client.
func NewVksSupervisorAuthClient(config VksAuthConfig) (*VksK8sAuthClient, error) {
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
	if _, lr, err := client.Login(); err != nil {
		return nil, err
	} else if lr.GuestClusterServer != "" && lr.GuestClusterCA != "" {
		client.cfg.Endpoint = "https://" + lr.GuestClusterServer + ":6443"
		client.tlsConfig.CAData = []byte(lr.GuestClusterCA)
	}

	client.tlsConfig, err = client.buildTLSConfig()
	if err != nil {
		return nil, fmt.Errorf("build TLS config failed: %w", err)
	}

	kubeConfig, err := client.buildVksKubeconfig()
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

// NewVksGuestClusterAuthClient creates a Kubernetes client using stored vSphere credentials for guest cluster
func NewVksGuestClusterAuthClient(config VksAuthConfig) (*VksK8sAuthClient, error) {

	// Validate the GuestClusterName is provided
	if config.GuestClusterName == "" {
		return nil, errors.New("guest cluster name is required")
	}
	vksAuthClient, err := NewVksSupervisorAuthClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create vSphere authenticated client: %w", err)
	}

	guestClusterClient, err := k8sapiClient.New(&rest.Config{
		Host:            vksAuthClient.cfg.Endpoint,
		BearerToken:     vksAuthClient.token,
		TLSClientConfig: vksAuthClient.tlsConfig,
	}, k8sapiClient.Options{})
	if err != nil {
		return nil, fmt.Errorf("create kubernetes client failed: %w", err)
	}

	vksAuthClient.Client = guestClusterClient
	return vksAuthClient, nil
}

// GetToken returns the JWT token stored in the VksK8sAuthClient struct.
func (c *VksK8sAuthClient) GetToken() string {
	return c.token
}

// Login performs the login to the VKS API server and stores the token in the VksK8sAuthClient struct.
// It returns the token, the loginResponse and any error encountered during the login process.
func (c *VksK8sAuthClient) Login() (token string, lr SupervisorLoginResponse, err error) {
	// Calls the login method to get the token and store it in c.token.
	token, lr, err = c.login()
	if err != nil {
		return "", lr, fmt.Errorf("login failed: %w, response: %s", err, lr)
	}
	c.token = token

	return c.token, lr, nil
}

// GenerateKubeconfig generates a kubeconfig string for the authenticated user to access the Kubernetes API server.
// It takes the cluster name and context name as parameters and returns the kubeconfig string and any error encountered during the process.
func (c *VksK8sAuthClient) GenerateKubeconfig(clusterName, contextName string) (kubeConfig string, err error) {

	// Validate the token before generating the kubeconfig.
	if valid, err := c.TokenValid(); !valid {
		return "", fmt.Errorf("token is not valid: %w", err)
	} else if err != nil {
		return "", fmt.Errorf("failed to validate token: %w", err)
	}

	// Generate the kubeconfig using the current configuration and token.
	kubeConfig, err = ConvertRESTConfigToKubeconfig(clusterName, c.cfg.Username, contextName, &rest.Config{
		Host:            c.cfg.Endpoint,
		BearerToken:     c.token,
		TLSClientConfig: c.tlsConfig,
	})

	return
}

func (c *VksK8sAuthClient) RefreshToken() (string, error) {
	// Perform login to refresh the token.
	token, _, err := c.Login()
	if err != nil {
		return "", fmt.Errorf("refresh token failed: %w", err)
	}

	return token, nil
}

// TokenExpiry returns the expiration time of the token stored in the VksK8sAuthClient.
//
// If the token is empty, malformed, or does not contain a valid "exp" claim,
// the zero value of time.Time is returned.
func (c *VksK8sAuthClient) TokenExpiry() time.Time {
	if c.token == "" {
		return time.Time{}
	}

	// Parse the JWT without validating the signature.
	// The client already obtained this token from the Supervisor authentication flow.
	// We only need to inspect the claims to determine the expiration time.
	token, _, err := new(jwt.Parser).ParseUnverified(
		c.token,
		jwt.MapClaims{},
	)
	if err != nil {
		return time.Time{}
	}

	// Extract the JWT claims.
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return time.Time{}
	}

	// Read the "exp" (expiration time) claim from the JWT.
	exp, err := claims.GetExpirationTime()
	if err != nil || exp == nil {
		return time.Time{}
	}

	// Return the expiration timestamp as a UTC time.Time.
	return exp.Time.UTC()
}

// TokenValid reports whether the token held by the VksK8sAuthClient is still valid.
//
// A return value of (true, nil) means the token exists and has not expired.
// A return value of (false, nil) means the token is validly parsed but expired.
//
// A non-nil error means the token could not be parsed or does not contain a valid expiration.
// (for example: empty token, malformed JWT, or missing expiration claim).
func (c *VksK8sAuthClient) TokenValid() (bool, error) {
	if c.token == "" {
		return false, errors.New("token is empty")
	}

	// Use TokenExpiry() as the single source of truth
	// for extracting the expiration timestamp.
	expiry := c.TokenExpiry()

	// A zero time means the expiration could not be determined.
	if expiry.IsZero() {
		return false, errors.New("token does not contain a valid expiration time")
	}

	// Compare the current time with the token expiration timestamp.
	return time.Now().UTC().Before(expiry), nil
}
