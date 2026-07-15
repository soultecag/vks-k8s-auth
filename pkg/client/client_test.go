package client_test

import (
	"testing"

	"github.com/soultecag/vks-k8s-auth/pkg/client"
)

func TestVksK8sAuthClient_GenerateKubeconfig(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		client *client.VksK8sAuthClient
		// Named input parameters for target function.
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
			// TODO: update the condition below to compare got with tt.want.
			if got != tt.want {
				t.Errorf("GenerateKubeconfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
