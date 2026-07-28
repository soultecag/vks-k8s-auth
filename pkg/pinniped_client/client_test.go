package client_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/soultecag/vks-k8s-auth/pkg/client"
)

func TestVksK8sAuthClient_GenerateKubeconfig(t *testing.T) {
	tests := []struct {
		name        string
		client      *client.VksK8sAuthClient
		clusterName string
		contextName string
		want        string
		wantErr     bool
	}{
		{
			name:        "no valid token",
			client:      &client.VksK8sAuthClient{},
			clusterName: "test-cluster",
			contextName: "test-context",
			want:        "",
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := tt.client.GenerateKubeconfig(tt.clusterName, tt.contextName)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GenerateKubeconfig() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GenerateKubeconfig() succeeded unexpectedly")
			}
			if got != tt.want {
				t.Errorf("GenerateKubeconfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVksK8sAuthClient_GetToken(t *testing.T) {
	tests := []struct {
		name   string
		client *client.VksK8sAuthClient
		want   string
	}{
		{
			name:   "zero value client has no token",
			client: &client.VksK8sAuthClient{},
			want:   "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.client.GetToken()
			if got != tt.want {
				t.Errorf("GetToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVksK8sAuthClient_TokenValid(t *testing.T) {
	tests := []struct {
		name    string
		client  *client.VksK8sAuthClient
		want    bool
		wantErr bool
	}{
		{
			name:    "empty token is invalid",
			client:  &client.VksK8sAuthClient{},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := tt.client.TokenValid()
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("TokenValid() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("TokenValid() succeeded unexpectedly")
			}
			if got != tt.want {
				t.Errorf("TokenValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVksK8sAuthClient_TokenExpiry(t *testing.T) {
	tests := []struct {
		name   string
		client *client.VksK8sAuthClient
		want   time.Time
	}{
		{
			name:   "empty token has zero expiry",
			client: &client.VksK8sAuthClient{},
			want:   time.Time{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.client.TokenExpiry()
			if !got.Equal(tt.want) {
				t.Errorf("TokenExpiry() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewVksSupervisorAuthClient(t *testing.T) {
	tests := []struct {
		name    string
		config  client.VksAuthConfig
		wantErr bool
	}{
		{
			name:    "empty endpoint",
			config:  client.VksAuthConfig{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := client.NewVksSupervisorAuthClient(tt.config)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("NewVksSupervisorAuthClient() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("NewVksSupervisorAuthClient() succeeded unexpectedly")
			}
			if got == nil {
				t.Errorf("NewVksSupervisorAuthClient() = nil, want non-nil client")
			}
		})
	}
}

func TestNewVksGuestClusterAuthClient(t *testing.T) {
	tests := []struct {
		name    string
		config  client.VksAuthConfig
		wantErr bool
	}{
		{
			name:    "missing guest cluster name and namespace",
			config:  client.VksAuthConfig{Endpoint: "https://10.0.0.1"},
			wantErr: true,
		},
		{
			name:    "missing guest cluster namespace",
			config:  client.VksAuthConfig{Endpoint: "https://10.0.0.1", GuestClusterName: "cluster"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := client.NewVksGuestClusterAuthClient(tt.config)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("NewVksGuestClusterAuthClient() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("NewVksGuestClusterAuthClient() succeeded unexpectedly")
			}
			if got == nil {
				t.Errorf("NewVksGuestClusterAuthClient() = nil, want non-nil client")
			}
		})
	}
}

func TestNewVksSupervisorAuthClient_Success(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != "admin" || pass != "secret" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"session_id": "integration-test-token"})
	}))
	defer server.Close()

	c, err := client.NewVksSupervisorAuthClient(client.VksAuthConfig{
		Endpoint: server.URL,
		Username: "admin",
		Password: "secret",
		// The test server uses a self-signed certificate. TlsInsecureSkipVerify
		// bypasses both the login TLS handshake and the CA-capture handshake,
		// since neither cert is trusted by the system store in this environment.
		TlsInsecureSkipVerify: true,
	})
	if err != nil {
		t.Fatalf("NewVksSupervisorAuthClient() failed: %v", err)
	}
	if c.GetToken() != "integration-test-token" {
		t.Errorf("GetToken() = %q, want %q", c.GetToken(), "integration-test-token")
	}
}

func TestNewVksSupervisorAuthClient_InvalidCredentials(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	_, err := client.NewVksSupervisorAuthClient(client.VksAuthConfig{
		Endpoint:              server.URL,
		Username:              "admin",
		Password:              "wrong",
		TlsInsecureSkipVerify: true,
	})
	if err == nil {
		t.Fatal("NewVksSupervisorAuthClient() succeeded unexpectedly")
	}
}

func TestVksK8sAuthClient_ResetHTTPClient_NoPanicWhenUninitialized(t *testing.T) {
	c := &client.VksK8sAuthClient{}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("ResetHTTPClient() panicked on a client that never made a request: %v", r)
		}
	}()

	c.ResetHTTPClient()
}

func TestVksK8sAuthClient_ResetHTTPClient_AfterUse(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"session_id": "token-abc"})
	}))
	defer server.Close()

	c, err := client.NewVksSupervisorAuthClient(client.VksAuthConfig{
		Endpoint:              server.URL,
		Username:              "admin",
		Password:              "secret",
		TlsInsecureSkipVerify: true,
	})
	if err != nil {
		t.Fatalf("NewVksSupervisorAuthClient() failed: %v", err)
	}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("ResetHTTPClient() panicked on an initialized client: %v", r)
		}
	}()

	c.ResetHTTPClient()

	// ResetHTTPClient only recycles the transport; it must not clear the token.
	if c.GetToken() != "token-abc" {
		t.Errorf("GetToken() after ResetHTTPClient() = %q, want %q", c.GetToken(), "token-abc")
	}
}
