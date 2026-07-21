package client_test

import (
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
