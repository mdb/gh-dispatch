package dispatch

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cli/go-gh/v2/pkg/api"
)

// newTestServer starts a TLS test server and returns a go-gh REST client
// that targets it, along with the mux on which to register API handlers.
// As the server's host is not github.com, go-gh treats it as a GitHub
// Enterprise Server host and prefixes all REST paths with /api/v3.
func newTestServer(t *testing.T) (*api.RESTClient, *http.ServeMux) {
	t.Helper()
	t.Setenv("GH_CONFIG_DIR", t.TempDir())

	mux := http.NewServeMux()
	srv := httptest.NewTLSServer(mux)
	t.Cleanup(srv.Close)

	client, err := api.NewRESTClient(api.ClientOptions{
		Host:      strings.TrimPrefix(srv.URL, "https://"),
		AuthToken: "token",
		Transport: srv.Client().Transport,
	})
	if err != nil {
		t.Fatal(err)
	}

	return client, mux
}

func respond(body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(body))
	}
}

func decodeBody(t *testing.T, r *http.Request) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		t.Errorf("failed to decode request body: %v", err)
	}
	return body
}
