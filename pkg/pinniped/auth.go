package pinniped

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

func loginOIDC(
	ctx context.Context,
	pinnipedURL string,
) (*oauth2.Token, error) {

	provider, err := oidc.NewProvider(
		ctx,
		pinnipedURL,
	)
	if err != nil {
		return nil, err
	}

	verifier := randomString(32)

	challenge := base64.RawURLEncoding.EncodeToString(
		[]byte(verifier),
	)

	redirect := "http://127.0.0.1:8080/callback"

	conf := oauth2.Config{
		ClientID: "pinniped-cli",

		Endpoint: provider.Endpoint(),

		RedirectURL: redirect,

		Scopes: []string{
			oidc.ScopeOpenID,
			"offline_access",
			"username",
			"groups",
		},
	}

	state := randomString(16)

	authURL := conf.AuthCodeURL(
		state,
		oauth2.SetAuthURLParam(
			"code_challenge",
			challenge,
		),
		oauth2.SetAuthURLParam(
			"code_challenge_method",
			"S256",
		),
	)

	code := make(chan string)

	server := &http.Server{
		Addr: "127.0.0.1:8080",
	}

	http.HandleFunc(
		"/callback",
		func(w http.ResponseWriter, r *http.Request) {

			if r.URL.Query().Get("state") != state {
				http.Error(
					w,
					"invalid state",
					400,
				)
				return
			}

			fmt.Fprint(
				w,
				"Login successful. You can close this window.",
			)

			code <- r.URL.Query().Get("code")
		},
	)

	go server.ListenAndServe()

	if err := openBrowser(authURL); err != nil {
		return nil, err
	}

	select {
	case c := <-code:

		token, err := conf.Exchange(
			ctx,
			c,
			oauth2.SetAuthURLParam(
				"code_verifier",
				verifier,
			),
		)

		server.Shutdown(ctx)

		return token, err

	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func randomString(size int) string {

	b := make([]byte, size)

	_, _ = rand.Read(b)

	return base64.RawURLEncoding.EncodeToString(b)
}

func openBrowser(url string) error {

	switch runtime.GOOS {

	case "linux":
		return exec.Command(
			"xdg-open",
			url,
		).Start()

	case "darwin":
		return exec.Command(
			"open",
			url,
		).Start()

	case "windows":
		return exec.Command(
			"rundll32",
			"url.dll,FileProtocolHandler",
			url,
		).Start()

	default:
		return fmt.Errorf(
			"cannot open browser",
		)
	}
}
