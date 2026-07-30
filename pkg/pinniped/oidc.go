package pinniped

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"k8s.io/client-go/rest"
)

const (
	defaultAuthenticatorName = "tkg-jwt-authenticator"
	defaultAuthenticatorKind = "JWTAuthenticator"
)

// Config contains the VKS Supervisor login configuration.
type Config struct {
	// Supervisor Pinniped endpoint.
	//
	// Example:
	// https://<supervisorIP>/wcp/pinniped
	PinnipedURL string

	// Guest cluster Kubernetes API endpoint.
	//
	// Example:
	// https://10.24.68.21:6443
	ClusterURL string

	// CA certificate of the Supervisor/Guest cluster.
	CABundle []byte

	// Optional. Defaults to tkg-jwt-authenticator.
	AuthenticatorName string

	// SupervisorConfig is the Supervisor configuration used to exchange the OIDC token for a Kubernetes token.
	// This is required for the corrected concierge.go.
	SupervisorConfig *rest.Config
}

// Login authenticates against VKS Pinniped and returns a Kubernetes rest.Config.
func Login(ctx context.Context, cfg Config) (*rest.Config, error) {
	if cfg.PinnipedURL == "" {
		return nil, fmt.Errorf("missing PinnipedURL")
	}

	if cfg.ClusterURL == "" {
		return nil, fmt.Errorf("missing ClusterURL")
	}

	authName := cfg.AuthenticatorName
	if authName == "" {
		authName = defaultAuthenticatorName
	}

	oidcToken, err := loginOIDC(ctx, cfg.PinnipedURL)
	if err != nil {
		return nil, fmt.Errorf("oidc login failed: %w", err)
	}

	token, err := exchangeToken(
		ctx,
		oidcToken,
		authName,
	)
	if err != nil {
		return nil, fmt.Errorf("concierge exchange failed: %w", err)
	}

	return &rest.Config{
		Host: cfg.ClusterURL,

		BearerToken: token,

		TLSClientConfig: rest.TLSClientConfig{
			CAData: cfg.CABundle,
		},
	}, nil
}

// EncodeCA is a helper if your CA is stored as *x509.Certificate.
func EncodeCA(cert *x509.Certificate) []byte {
	return pem.EncodeToMemory(
		&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: cert.Raw,
		},
	)
}
