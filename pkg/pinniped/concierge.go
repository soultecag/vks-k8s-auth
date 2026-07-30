package pinniped

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/oauth2"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	loginv1alpha1 "go.pinniped.dev/generated/latest/apis/concierge/login/v1alpha1"
	// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	conciergeAPIPath = "/apis/authentication.concierge.pinniped.dev/v1alpha1/tokencredentialrequests"

	authenticatorGroup = "authentication.concierge.pinniped.dev"
)

func exchangeToken(
	ctx context.Context,
	oidcToken *oauth2.Token,
	authenticatorName string,
) (string, error) {

	if oidcToken == nil {
		return "", fmt.Errorf("missing oidc token")
	}

	const authenticatorGroup = "authentication.concierge.pinniped.dev"

	req := loginv1alpha1.TokenCredentialRequest{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "login.concierge.pinniped.dev/v1alpha1",
			Kind:       "TokenCredentialRequest",
		},

		Spec: loginv1alpha1.TokenCredentialRequestSpec{
			Token: oidcToken.AccessToken,

			Authenticator: corev1.TypedLocalObjectReference{
				APIGroup: ptr(authenticatorGroup),
				Kind:     "JWTAuthenticator",
				Name:     authenticatorName,
			},
		},
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	// The Concierge endpoint is served by the Supervisor Kubernetes API.
	//
	// In VKS this is normally:
	//
	// https://<supervisor>/apis/authentication.concierge.pinniped.dev/v1alpha1/tokencredentialrequests
	//
	// The issuer URL and Kubernetes API URL are the same Supervisor endpoint.
	//
	// Therefore we derive it from the issuer.

	supervisor := strings.TrimSuffix(
		strings.TrimSuffix(
			oidcToken.Extra("issuer").(string),
			"/wcp/pinniped",
		),
		"/",
	)

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		supervisor+conciergeAPIPath,
		strings.NewReader(string(body)),
	)

	if err != nil {
		return "", err
	}

	httpReq.Header.Set(
		"Content-Type",
		"application/json",
	)

	client := http.DefaultClient

	resp, err := client.Do(httpReq)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return "",
			fmt.Errorf(
				"concierge returned %s",
				resp.Status,
			)
	}

	var response loginv1alpha1.TokenCredentialRequest

	if err := json.NewDecoder(
		resp.Body,
	).Decode(&response); err != nil {
		return "", err
	}

	if response.Status.Credential == nil {
		return "",
			fmt.Errorf(
				"concierge returned no credential",
			)
	}

	return response.Status.Credential.Token, nil
}

func ptr[T any](v T) *T {
	return &v
}
