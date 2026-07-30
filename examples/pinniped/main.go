package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"k8s.io/client-go/kubernetes"

	vksauth "github.com/soultecag/vks-k8s-auth/pkg/pinniped"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// "k8s.io/client-go/kubernetes"
)

type Discovery struct {
	Issuer                string `json:"issuer"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	JWKSURI               string `json:"jwks_uri"`
}

func main() {
	ctx := context.Background()

	req, _ := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://10.24.68.5/wcp/pinniped/.well-known/openid-configuration",
		nil,
	)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var d Discovery

	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		panic(err)
	}

	fmt.Printf("Issuer: %s\n", d.Issuer)
	fmt.Printf("Auth:   %s\n", d.AuthorizationEndpoint)
	fmt.Printf("Token:  %s\n", d.TokenEndpoint)

	// VKS Guest cluster API endpoint.
	//
	// Get it from:
	//
	// kubectl get clusters -A
	//
	// or from the VKS API endpoint.
	//
	guestClusterAPI := "https://10.24.68.21:6443"

	// Supervisor Pinniped endpoint.
	pinniped := "https://10.24.68.5/wcp/pinniped"

	cfg, err := vksauth.Login(
		ctx,
		vksauth.Config{
			PinnipedURL: pinniped,

			ClusterURL: guestClusterAPI,

			// From:
			//
			// kubectl get jwtauthenticators.authentication.concierge.pinniped.dev
			//
			AuthenticatorName: "tkg-jwt-authenticator",
		},
	)

	if err != nil {
		panic(err)
	}

	k8sClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		panic(err)
	}

	nodes, err := k8sClient.CoreV1().
		Nodes().
		List(
			ctx,
			metav1.ListOptions{},
		)

	if err != nil {
		panic(err)
	}

	fmt.Printf(
		"Guest cluster has %d nodes\n",
		len(nodes.Items),
	)

}
