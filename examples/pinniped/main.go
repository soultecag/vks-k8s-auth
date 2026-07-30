package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

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

var (
	// port         int
	supervisorIP string
	// username     string
	// password     string

	// VKS Guest cluster API endpoint.
	//
	// Get it from:
	//
	// kubectl get clusters -A
	//
	// or from the VKS API endpoint.
	//
	guestClusterAPI string
)

func init() {

	//Read the VKS API server endpoint, username, and password from environment variables (SUPERVISOR_ENDPOINT, VSPHERE_USERNAME and VSPHERE_PASSWORD )

	supervisorIP = os.Getenv("SUPERVISOR_ENDPOINT")
	guestClusterAPI = os.Getenv("GUEST_CLUSTER_API")
	// username = os.Getenv("VSPHERE_USERNAME")
	// password = os.Getenv("VSPHERE_PASSWORD")
	// portString := os.Getenv("VSPHERE_PORT")

	if supervisorIP == "" || guestClusterAPI == "" {
		panic("SUPERVISOR_ENDPOINT and GUEST_CLUSTER_API environment variables must be set, e.g. SUPERVISOR_ENDPOINT=10.24.68.5 GUEST_CLUSTER_API=https://1.2.3.4:6443")
	}

	// if portString != "" {
	// 	parsedPort, err := strconv.Atoi(portString)
	// 	if err == nil {
	// 		port = parsedPort
	// 	}
	// }

}

func main() {
	ctx := context.Background()

	req, _ := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("https://%s/wcp/pinniped/.well-known/openid-configuration", supervisorIP),
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

	// Supervisor Pinniped endpoint.
	pinniped := fmt.Sprintf("https://%s/wcp/pinniped", supervisorIP)

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
