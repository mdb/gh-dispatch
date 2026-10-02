package dispatch

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/stretchr/testify/assert"
)

func TestRepositoryDispatchRun(t *testing.T) {
	ghRepo := &ghRepo{
		Name:  "REPO",
		Owner: "OWNER",
	}
	repo := ghRepo.RepoFullName()
	event := "repository_dispatch"
	prefix := fmt.Sprintf("/api/v3/repos/%s", repo)

	registerHandlers := func(t *testing.T, mux *http.ServeMux, conclusion, jobsResponse string) {
		mux.HandleFunc("POST "+prefix+"/dispatches", func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, map[string]any{
				"event_type":     "hello",
				"client_payload": any(nil),
			}, decodeBody(t, r))
			w.WriteHeader(http.StatusNoContent)
		})

		mux.HandleFunc("GET "+prefix+"/actions/workflows", func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "100", r.URL.Query().Get("per_page"))
			assert.Equal(t, "1", r.URL.Query().Get("page"))
			respond(getWorkflowsResponse)(w, r)
		})

		mux.HandleFunc("GET /api/v3/user", respond(currentUserResponse))

		mux.HandleFunc("GET "+prefix+"/actions/workflows/456/runs", func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "mdb", r.URL.Query().Get("actor"))
			respond(fmt.Sprintf(getWorkflowRunsResponse, event, repo))(w, r)
		})

		mux.HandleFunc("GET "+prefix+"/actions/runs/123", func(w http.ResponseWriter, r *http.Request) {
			respond(fmt.Sprintf(`{
				"id": 123,
				"workflow_id": 456,
				"status": "completed",
				"event": "repository_dispatch",
				"conclusion": "%s",
				"jobs_url": "https://%s%s/actions/runs/123/jobs"
			}`, conclusion, r.Host, prefix))(w, r)
		})

		mux.HandleFunc("GET "+prefix+"/actions/workflows/456", respond(getWorkflowResponse))
		mux.HandleFunc("GET "+prefix+"/actions/runs/123/jobs", respond(jobsResponse))
		mux.HandleFunc("GET "+prefix+"/check-runs/123/annotations", respond("[]"))
	}

	tests := []struct {
		name     string
		opts     *repositoryDispatchOptions
		handlers func(*testing.T, *http.ServeMux)
		wantErr  bool
		errMsg   string
		wantOut  string
	}{
		{
			name: "successful workflow run",
			opts: &repositoryDispatchOptions{
				eventType: "hello",
				workflow:  "foo",
			},
			handlers: func(t *testing.T, mux *http.ServeMux) {
				registerHandlers(t, mux, "success", getJobsResponse)
			},
			wantOut: `Refreshing run status every 2 seconds. Press Ctrl+C to quit.

https://github.com/OWNER/REPO/actions/runs/123

✓  foo · 123
Triggered via repository_dispatch 

JOBS
✓ build in 1m59s (ID 123)
  ✓ Run actions/checkout@v2
  ✓ Test
`,
		}, {
			name: "unsuccessful workflow run",
			opts: &repositoryDispatchOptions{
				eventType: "hello",
				workflow:  "foo",
			},
			handlers: func(t *testing.T, mux *http.ServeMux) {
				registerHandlers(t, mux, "failure", getFailingJobsResponse)
			},
			wantOut: `Refreshing run status every 2 seconds. Press Ctrl+C to quit.

https://github.com/OWNER/REPO/actions/runs/123

X  foo · 123
Triggered via repository_dispatch 

JOBS
✓ build in 1m59s (ID 123)
  ✓ Run actions/checkout@v2
  X Test
`,
			wantErr: true,
			errMsg:  "SilentError",
		}, {
			name: "malformed JSON response",
			opts: &repositoryDispatchOptions{
				eventType: "hello",
			},
			handlers: func(t *testing.T, mux *http.ServeMux) {
				mux.HandleFunc("POST "+prefix+"/dispatches", respond("{"))
			},
			wantOut: "",
			wantErr: true,
			errMsg:  "unexpected end of JSON input",
		}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mux := newTestServer(t)
			tt.handlers(t, mux)

			ios, _, stdout, _ := iostreams.Test()
			ios.SetStdoutTTY(false)
			ios.SetAlternateScreenBufferEnabled(false)

			tt.opts.repo = ghRepo
			tt.opts.io = ios
			tt.opts.client = client

			err := repositoryDispatchRun(tt.opts)

			if tt.wantErr {
				assert.EqualError(t, err, tt.errMsg)
			} else {
				assert.NoError(t, err)
			}

			if got := stdout.String(); got != tt.wantOut {
				t.Errorf("got stdout:\n%q\nwant:\n%q", got, tt.wantOut)
			}
		})
	}
}
