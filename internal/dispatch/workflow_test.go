package dispatch

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/cli/cli/v2/pkg/httpmock"
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

	createMockRegistry := func(reg *httpmock.Registry, conclusion, jobsResponse string) {
		reg.Register(
			httpmock.REST("POST", fmt.Sprintf("repos/%s/actions/workflows/%s/dispatches", repo, "workflow.yaml")),
			httpmock.RESTPayload(201, "{}", func(params map[string]interface{}) {
				assert.Equal(t, map[string]interface{}{
					"inputs": "{\"foo\": \"bar\"}",
					"ref":    "",
				}, params)
			}))

		reg.Register(
			httpmock.REST("GET", fmt.Sprintf("repos/%s/actions/workflows/workflow.yaml", repo)),
			httpmock.StringResponse(getWorkflowResponse))

		reg.Register(
			httpmock.GraphQL("query UserCurrent{viewer{login}}"),
			httpmock.StringResponse(currentUserResponse))

		v := url.Values{}
		v.Set("per_page", "50")

		reg.Register(
			httpmock.QueryMatcher("GET", fmt.Sprintf("repos/%s/actions/workflows/456/runs", repo), v),
			httpmock.StringResponse(fmt.Sprintf(getWorkflowRunsResponse, event)))

		q := url.Values{}
		q.Set("per_page", "100")
		q.Set("page", "1")

		reg.Register(
			httpmock.QueryMatcher("GET", fmt.Sprintf("repos/%s/actions/workflows", repo), q),
			httpmock.StringResponse(fmt.Sprintf(getWorkflowRunsResponse, event)))

		reg.Register(
			httpmock.REST("GET", fmt.Sprintf("repos/%s/actions/workflows/456", repo)),
			httpmock.StringResponse(getWorkflowResponse))

		reg.Register(
			httpmock.REST("GET", fmt.Sprintf("repos/%s/actions/runs/123", repo)),
			httpmock.StringResponse(`{
				"id": 123,
				"workflow_id": 456,
				"event": "workflow_dispatch"
			}`))

		reg.Register(
			httpmock.REST("GET", fmt.Sprintf("repos/%s/actions/workflows/456", repo)),
			httpmock.StringResponse(getWorkflowResponse))

		reg.Register(
			httpmock.REST("GET", fmt.Sprintf("repos/%s/actions/workflows/456", repo)),
			httpmock.StringResponse(getWorkflowResponse))

		reg.Register(
			httpmock.REST("GET", fmt.Sprintf("repos/%s/actions/runs/123", repo)),
			httpmock.StringResponse(fmt.Sprintf(`{
				"id": 123,
				"workflow_id": 456,
				"event": "workflow_dispatch",
				"status": "completed",
				"conclusion": "%s",
				"jobs_url": "https://api.github.com/repos/%s/actions/runs/123/jobs"
			}`, conclusion, repo)))

		reg.Register(
			httpmock.REST("GET", fmt.Sprintf("repos/%s/actions/runs/123/jobs", repo)),
			httpmock.StringResponse(jobsResponse))

		reg.Register(
			httpmock.REST("GET", fmt.Sprintf("repos/%s/check-runs/123/annotations", repo)),
			httpmock.StringResponse("[]"))
	}

	tests := []struct {
		name      string
		opts      *workflowDispatchOptions
		httpStubs func(*httpmock.Registry)
		wantErr   bool
		errMsg    string
		wantOut   string
	}{
		{
			name: "successful workflow run",
			opts: &workflowDispatchOptions{
				inputs:   `{"foo": "bar"}`,
				workflow: workflow,
			},
			httpStubs: func(reg *httpmock.Registry) {
				createMockRegistry(reg, "success", getJobsResponse)
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
			httpStubs: func(reg *httpmock.Registry) {
				createMockRegistry(reg, "failure", getFailingJobsResponse)
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
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("POST", fmt.Sprintf("repos/%s/actions/workflows/%s/dispatches", repo, "workflow.yaml")),
					httpmock.StringResponse("{"))
			},
			wantOut: "",
			wantErr: true,
			errMsg:  "unexpected end of JSON input",
		}}

	for _, tt := range tests {
		reg := &httpmock.Registry{}
		tt.httpStubs(reg)

		ios, _, stdout, _ := iostreams.Test()
		ios.SetStdoutTTY(false)
		ios.SetAlternateScreenBufferEnabled(false)

		tt.opts.repo = ghRepo
		tt.opts.io = ios
		tt.opts.httpClient = &http.Client{
			Transport: reg,
		}

		t.Run(tt.name, func(t *testing.T) {
			err := workflowDispatchRun(tt.opts)

			if tt.wantErr {
				assert.EqualError(t, err, tt.errMsg)
			} else {
				assert.NoError(t, err)
			}

			if got := stdout.String(); got != tt.wantOut {
				t.Errorf("got stdout:\n%q\nwant:\n%q", got, tt.wantOut)
			}

			reg.Verify(t)
		})
	}
}
