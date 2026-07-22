package client

import (
	"crypto/x509"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// --- getSupervisorHost -------------------------------------------------

func TestGetSupervisorHost(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		port     int
		want     string
		wantErr  bool
	}{
		{name: "empty endpoint errors", endpoint: "", port: 0, wantErr: true},
		{name: "whitespace only errors", endpoint: "   ", port: 0, wantErr: true},
		{name: "bare host gets https prefix", endpoint: "10.5.24.5", port: 0, want: "https://10.5.24.5"},
		{name: "http prefix preserved", endpoint: "http://10.5.24.5", port: 0, want: "http://10.5.24.5"},
		{name: "https prefix preserved", endpoint: "https://10.5.24.5", port: 0, want: "https://10.5.24.5"},
		{name: "trailing slash trimmed", endpoint: "https://10.5.24.5/", port: 0, want: "https://10.5.24.5"},
		{name: "port appended when non-zero", endpoint: "https://10.5.24.5", port: 6443, want: "https://10.5.24.5:6443"},
		{name: "surrounding whitespace trimmed", endpoint: "  10.5.24.5  ", port: 0, want: "https://10.5.24.5"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getSupervisorHost(tt.endpoint, tt.port)
			if tt.wantErr {
				if err == nil {
					t.Fatal("getSupervisorHost() succeeded unexpectedly")
				}
				return
			}
			if err != nil {
				t.Fatalf("getSupervisorHost() failed: %v", err)
			}
			if got != tt.want {
				t.Errorf("getSupervisorHost() = %q, want %q", got, tt.want)
			}
		})
	}
}

// --- newHTTPClient ------------------------------------------------------

func TestNewHTTPClient(t *testing.T) {
	t.Run("default timeout applied when unset", func(t *testing.T) {
		hc := newHTTPClient(VksAuthConfig{})
		if hc.Timeout != defaultTimeoutSeconds*time.Second {
			t.Errorf("Timeout = %v, want %v", hc.Timeout, defaultTimeoutSeconds*time.Second)
		}
	})

	t.Run("default timeout applied when negative", func(t *testing.T) {
		hc := newHTTPClient(VksAuthConfig{Timeout: -5})
		if hc.Timeout != defaultTimeoutSeconds*time.Second {
			t.Errorf("Timeout = %v, want %v", hc.Timeout, defaultTimeoutSeconds*time.Second)
		}
	})

	t.Run("custom timeout honored", func(t *testing.T) {
		hc := newHTTPClient(VksAuthConfig{Timeout: 5})
		if hc.Timeout != 5*time.Second {
			t.Errorf("Timeout = %v, want %v", hc.Timeout, 5*time.Second)
		}
	})

	t.Run("insecure skip verify propagated to transport", func(t *testing.T) {
		hc := newHTTPClient(VksAuthConfig{TlsInsecureSkipVerify: true})
		transport, ok := hc.Transport.(*http.Transport)
		if !ok {
			t.Fatalf("Transport is not *http.Transport: %T", hc.Transport)
		}
		if !transport.TLSClientConfig.InsecureSkipVerify {
			t.Error("InsecureSkipVerify = false, want true")
		}
	})
}

// --- supervisorDialAddress ----------------------------------------------

func TestSupervisorDialAddress(t *testing.T) {
	tests := []struct {
		name    string
		server  string
		want    string
		wantErr bool
	}{
		{name: "explicit port preserved", server: "https://10.5.24.5:6443", want: "10.5.24.5:6443"},
		{name: "defaults to port 443", server: "https://10.5.24.5", want: "10.5.24.5:443"},
		{name: "empty host errors", server: "https://", wantErr: true},
		{name: "unparseable url errors", server: "://bad-url", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := supervisorDialAddress(tt.server)
			if tt.wantErr {
				if err == nil {
					t.Fatal("supervisorDialAddress() succeeded unexpectedly")
				}
				return
			}
			if err != nil {
				t.Fatalf("supervisorDialAddress() failed: %v", err)
			}
			if got != tt.want {
				t.Errorf("supervisorDialAddress() = %q, want %q", got, tt.want)
			}
		})
	}
}

// --- pickCA ---------------------------------------------------------------

func TestPickCA(t *testing.T) {
	leaf := &x509.Certificate{Raw: []byte("leaf")}
	intermediate := &x509.Certificate{Raw: []byte("intermediate"), IsCA: true}
	root := &x509.Certificate{Raw: []byte("root"), IsCA: true}

	t.Run("returns the CA closest to the root", func(t *testing.T) {
		got := pickCA([]*x509.Certificate{leaf, intermediate, root})
		if got != root {
			t.Error("pickCA() did not return the root CA")
		}
	})

	t.Run("falls back to sole leaf when no CA present (typical k8s API server chain)", func(t *testing.T) {
		got := pickCA([]*x509.Certificate{leaf})
		if got != leaf {
			t.Error("pickCA() did not fall back to the leaf certificate")
		}
	})

	t.Run("picks the available CA when no root is present", func(t *testing.T) {
		got := pickCA([]*x509.Certificate{leaf, intermediate})
		if got != intermediate {
			t.Error("pickCA() did not return the intermediate CA")
		}
	})
}

// --- ensureHTTPClient -----------------------------------------------------

func TestEnsureHTTPClient_Singleton(t *testing.T) {
	c := &VksK8sAuthClient{cfg: VksAuthConfig{}}

	first := c.ensureHTTPClient()
	second := c.ensureHTTPClient()
	if first != second {
		t.Error("ensureHTTPClient() returned different instances on repeated calls")
	}
}

func TestEnsureHTTPClient_ConcurrentSingleton(t *testing.T) {
	c := &VksK8sAuthClient{cfg: VksAuthConfig{}}

	const goroutines = 50
	results := make(chan *http.Client, goroutines)
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			results <- c.ensureHTTPClient()
		}()
	}
	wg.Wait()
	close(results)

	var first *http.Client
	for hc := range results {
		if first == nil {
			first = hc
			continue
		}
		if hc != first {
			t.Fatal("ensureHTTPClient() returned different instances under concurrent access")
		}
	}
}

// --- buildTLSConfig / buildVksKubeconfig -----------------------------------

func TestBuildTLSConfig_InsecureSkipVerify(t *testing.T) {
	c := &VksK8sAuthClient{cfg: VksAuthConfig{TlsInsecureSkipVerify: true}}
	tlsCfg, err := c.buildTLSConfig()
	if err != nil {
		t.Fatalf("buildTLSConfig() failed: %v", err)
	}
	if !tlsCfg.Insecure {
		t.Error("Insecure = false, want true")
	}
	if len(tlsCfg.CAData) != 0 {
		t.Error("CAData should be empty when TlsInsecureSkipVerify is true")
	}
}

func TestBuildVksKubeconfig(t *testing.T) {
	t.Run("empty token errors", func(t *testing.T) {
		c := &VksK8sAuthClient{}
		if _, err := c.buildVksKubeconfig(); err == nil {
			t.Fatal("buildVksKubeconfig() succeeded unexpectedly")
		}
	})

	t.Run("populates rest.Config from client state", func(t *testing.T) {
		c := &VksK8sAuthClient{
			token: "abc123",
			cfg:   VksAuthConfig{Endpoint: "https://10.5.24.5"},
		}
		cfg, err := c.buildVksKubeconfig()
		if err != nil {
			t.Fatalf("buildVksKubeconfig() failed: %v", err)
		}
		if cfg.Host != "https://10.5.24.5" {
			t.Errorf("Host = %q, want %q", cfg.Host, "https://10.5.24.5")
		}
		if cfg.BearerToken != "abc123" {
			t.Errorf("BearerToken = %q, want %q", cfg.BearerToken, "abc123")
		}
	})
}

// --- login ------------------------------------------------------------------

func TestLogin(t *testing.T) {
	tests := []struct {
		name      string
		handler   http.HandlerFunc
		username  string
		password  string
		wantToken string
		wantErr   bool
	}{
		{
			name: "successful login returns session token",
			handler: func(w http.ResponseWriter, r *http.Request) {
				user, pass, ok := r.BasicAuth()
				if !ok || user != "admin" || pass != "secret" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				_ = json.NewEncoder(w).Encode(SupervisorLoginResponse{SessionID: "test-session-token"})
			},
			username:  "admin",
			password:  "secret",
			wantToken: "test-session-token",
		},
		{
			name: "invalid credentials rejected",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte("unauthorized"))
			},
			username: "admin",
			password: "wrong",
			wantErr:  true,
		},
		{
			name: "missing session id in response errors",
			handler: func(w http.ResponseWriter, r *http.Request) {
				_ = json.NewEncoder(w).Encode(SupervisorLoginResponse{})
			},
			username: "admin",
			password: "secret",
			wantErr:  true,
		},
		{
			name: "malformed json response errors",
			handler: func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte("not-json"))
			},
			username: "admin",
			password: "secret",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			c := &VksK8sAuthClient{cfg: VksAuthConfig{
				Endpoint: server.URL,
				Username: tt.username,
				Password: tt.password,
			}}

			token, _, err := c.login()
			if tt.wantErr {
				if err == nil {
					t.Fatal("login() succeeded unexpectedly")
				}
				return
			}
			if err != nil {
				t.Fatalf("login() failed: %v", err)
			}
			if token != tt.wantToken {
				t.Errorf("login() = %q, want %q", token, tt.wantToken)
			}
		})
	}
}

func TestLogin_SendsExpectedRequest(t *testing.T) {
	var gotUser, gotPass, gotContentType string
	var gotBody SupervisorLoginRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUser, gotPass, _ = r.BasicAuth()
		gotContentType = r.Header.Get("Content-Type")
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_ = json.NewEncoder(w).Encode(SupervisorLoginResponse{SessionID: "token-123"})
	}))
	defer server.Close()

	c := &VksK8sAuthClient{cfg: VksAuthConfig{
		Endpoint:              server.URL,
		Username:              "admin",
		Password:              "secret",
		GuestClusterName:      "my-cluster",
		GuestClusterNamespace: "my-namespace",
	}}

	if _, _, err := c.login(); err != nil {
		t.Fatalf("login() failed: %v", err)
	}

	if gotUser != "admin" || gotPass != "secret" {
		t.Errorf("BasicAuth = (%q, %q), want (%q, %q)", gotUser, gotPass, "admin", "secret")
	}
	if gotContentType != "application/json" {
		t.Errorf("Content-Type = %q, want %q", gotContentType, "application/json")
	}
	if gotBody.GuestClusterName != "my-cluster" || gotBody.GuestClusterNamespace != "my-namespace" {
		t.Errorf("request body = %+v, want GuestClusterName=my-cluster GuestClusterNamespace=my-namespace", gotBody)
	}
}

// --- TokenExpiry / TokenValid with real (unverified) JWTs -------------------

func newTestJWT(t *testing.T, exp time.Time) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": exp.Unix(),
	})
	signed, err := tok.SignedString([]byte("test-signing-key"))
	if err != nil {
		t.Fatalf("failed to sign test token: %v", err)
	}
	return signed
}

func TestTokenExpiry_RealTokens(t *testing.T) {
	future := time.Now().Add(time.Hour).Truncate(time.Second)
	past := time.Now().Add(-time.Hour).Truncate(time.Second)

	tests := []struct {
		name  string
		token string
		want  time.Time
	}{
		{name: "empty token", token: "", want: time.Time{}},
		{name: "malformed token", token: "not-a-jwt", want: time.Time{}},
		{name: "valid token with future expiry", token: newTestJWT(t, future), want: future.UTC()},
		{name: "valid token with past expiry", token: newTestJWT(t, past), want: past.UTC()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &VksK8sAuthClient{token: tt.token}
			got := c.TokenExpiry()
			if !got.Equal(tt.want) {
				t.Errorf("TokenExpiry() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTokenValid_RealTokens(t *testing.T) {
	future := time.Now().Add(time.Hour)
	past := time.Now().Add(-time.Hour)

	t.Run("valid unexpired token", func(t *testing.T) {
		c := &VksK8sAuthClient{token: newTestJWT(t, future)}
		valid, err := c.TokenValid()
		if err != nil {
			t.Fatalf("TokenValid() failed: %v", err)
		}
		if !valid {
			t.Error("TokenValid() = false, want true")
		}
	})

	t.Run("expired token is invalid without error", func(t *testing.T) {
		c := &VksK8sAuthClient{token: newTestJWT(t, past)}
		valid, err := c.TokenValid()
		if err != nil {
			t.Fatalf("TokenValid() failed unexpectedly: %v", err)
		}
		if valid {
			t.Error("TokenValid() = true, want false for expired token")
		}
	})

	t.Run("malformed token errors", func(t *testing.T) {
		c := &VksK8sAuthClient{token: "not-a-jwt"}
		if _, err := c.TokenValid(); err == nil {
			t.Fatal("TokenValid() succeeded unexpectedly")
		}
	})
}
