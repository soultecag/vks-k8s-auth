package client

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"k8s.io/client-go/rest"
)

// buildSupervisorKubeconfig creates a REST config for vSphere Kubernetes authentication
// This connects to the supervisorCLuster API server using the provided endpoint, bearer token.
func (c *VksK8sAuthClient) buildSupervisorKubeconfig() (*rest.Config, error) {
	if c.token == "" {
		return nil, fmt.Errorf("bearer token is required")
	}
	return &rest.Config{
		Host:            c.cfg.Endpoint,
		BearerToken:     c.token,
		TLSClientConfig: c.tlsConfig,
	}, nil
}

// getSupervisorHost validates and formats the supervisor endpoint URL (e.g. https://10.5.24.5)
// if no protocol is provided, it defaults to https://. It also trims any trailing slashes.
func getSupervisorHost(supervisorEndpoint string, port int) (string, error) {
	host := strings.TrimSpace(supervisorEndpoint)
	if host == "" {
		return "", fmt.Errorf("supervisor endpoint is required")
	}

	if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
		host = "https://" + host
	}
	host = strings.TrimRight(host, "/")
	if port != 0 {
		host = fmt.Sprintf("%s:%d", host, port)
	}
	return host, nil
}

// login POSTs to /wcp/login with Basic auth and returns the session token.
func (c *VksK8sAuthClient) login() (token, raw string, err error) {

	url := fmt.Sprintf("%s/wcp/login", c.cfg.Endpoint)

	tlsconfig := &tls.Config{
		InsecureSkipVerify: c.cfg.TlsInsecureSkipVerify,
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: tlsconfig,
		},
	}

	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader("{}"))
	if err != nil {
		return "", "", err
	}
	req.SetBasicAuth(c.cfg.Username, c.cfg.Password)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	raw = string(body)

	if resp.StatusCode != http.StatusOK {
		return "", raw, fmt.Errorf("unexpected status %s", resp.Status)
	}

	var lr struct {
		SessionID string `json:"session_id"`
	}
	if err := json.Unmarshal(body, &lr); err != nil {
		return "", raw, fmt.Errorf("decode response: %w", err)
	}
	if lr.SessionID == "" {
		return "", raw, fmt.Errorf("no session_id in response")
	}
	return lr.SessionID, raw, nil
}
func (c *VksK8sAuthClient) buildTLSConfig() (rest.TLSClientConfig, error) {

	if !c.cfg.TlsInsecureSkipVerify {
		caPEM, err := c.fetchServerCA()
		if err != nil {
			fmt.Fprintf(os.Stderr, "warn: could not capture API server CA (%v); using system trust store\n", err)

		}
		return rest.TLSClientConfig{
			CAData: caPEM,
		}, nil
	}

	return rest.TLSClientConfig{
		Insecure: c.cfg.TlsInsecureSkipVerify,
	}, nil
}

// fetchServerCA does a TLS handshake against the API server and returns the
// PEM-encoded root of the presented chain (the Supervisor CA).
func (c *VksK8sAuthClient) fetchServerCA() ([]byte, error) {
	addr, err := supervisorDialAddress(c.cfg.Endpoint)
	if err != nil {
		return nil, err
	}

	conn, err := tls.Dial("tcp", addr, &tls.Config{InsecureSkipVerify: c.cfg.TlsInsecureSkipVerify})
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	chain := conn.ConnectionState().PeerCertificates
	if len(chain) == 0 {
		return nil, fmt.Errorf("no peer certificates")
	}
	// Prefer the last cert in the chain (the CA/root). If a single leaf is
	// presented, use it directly.
	ca := pickCA(chain)
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: ca.Raw}), nil
}

func supervisorDialAddress(server string) (string, error) {
	parsed, err := url.Parse(server)
	if err != nil {
		return "", fmt.Errorf("parse supervisor endpoint: %w", err)
	}

	host := parsed.Hostname()
	if host == "" {
		host = parsed.Path
	}
	if host == "" {
		return "", fmt.Errorf("supervisor endpoint host is required")
	}

	port := parsed.Port()
	if port == "" {
		port = "443"
	}

	return net.JoinHostPort(host, port), nil
}

func pickCA(chain []*x509.Certificate) *x509.Certificate {
	for i := len(chain) - 1; i >= 0; i-- {
		if chain[i].IsCA {
			return chain[i]
		}
	}
	return chain[len(chain)-1]
}
