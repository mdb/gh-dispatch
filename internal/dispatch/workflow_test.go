package dispatch

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/stretchr/testify/assert"
)

func TestWorkflowDispatchRun(t *testing.T) {
	ghRepo := &ghRepo{
		Owner: "OWNER",
		Name:  "REPO",
	}
	repo := ghRepo.RepoFullName()
	workflow := "workflow.yaml"
	event := "workflow_dispatch"
	prefix := fmt.Sprintf("/api/v3/repos/%s", repo)

	registerHandlers := func(t *testing.T, mux *http.ServeMux, conclusion, jobsResponse string) {
		mux.HandleFunc("POST "+prefix+"/actions/workflows/"+workflow+"/dispatches", func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, map[string]any{
				"inputs": "{\"foo\": \"bar\"}",
				"ref":    "",
			}, decodeBody(t, r))
			w.WriteHeader(http.StatusNoContent)
		})

		mux.HandleFunc("GET "+prefix+"/actions/workflows/"+workflow, respond(getWorkflowResponse))

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
				"event": "workflow_dispatch",
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
		opts     *workflowDispatchOptions
		handlers func(*testing.T, *http.ServeMux)
		wantErr  bool
		errMsg   string
		wantOut  string
	}{
		{
			name: "successful workflow run",
			opts: &workflowDispatchOptions{
				inputs:   `{"foo": "bar"}`,
				workflow: workflow,
			},
			handlers: func(t *testing.T, mux *http.ServeMux) {
				registerHandlers(t, mux, "success", getJobsResponse)
			},
			wantOut: `Refreshing run status every 2 seconds. Press Ctrl+C to quit.

https://github.com/OWNER/REPO/actions/runs/123

✓  foo · 123
Triggered via workflow_dispatch 

JOBS
✓ build in 1m59s (ID 123)
  ✓ Run actions/checkout@v2
  ✓ Test
`,
		}, {
			name: "unsuccessful workflow run",
			opts: &workflowDispatchOptions{
				inputs:   `{"foo": "bar"}`,
				workflow: workflow,
			},
			handlers: func(t *testing.T, mux *http.ServeMux) {
				registerHandlers(t, mux, "failure", getFailingJobsResponse)
			},
			wantOut: `Refreshing run status every 2 seconds. Press Ctrl+C to quit.

https://github.com/OWNER/REPO/actions/runs/123

X  foo · 123
Triggered via workflow_dispatch 

JOBS
✓ build in 1m59s (ID 123)
  ✓ Run actions/checkout@v2
  X Test
`,
			wantErr: true,
			errMsg:  "SilentError",
		}, {
			name: "malformed JSON response",
			opts: &workflowDispatchOptions{
				inputs:   `{"foo": "bar"}`,
				workflow: workflow,
			},
			handlers: func(t *testing.T, mux *http.ServeMux) {
				mux.HandleFunc("POST "+prefix+"/actions/workflows/"+workflow+"/dispatches", respond("{"))
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

			err := workflowDispatchRun(tt.opts)

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
