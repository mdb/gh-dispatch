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
				// TODO: test that proper request body is sent on POSTs
				reg.Register(
					httpmock.REST("POST", fmt.Sprintf("repos/%s/actions/workflows/%s/dispatches", repo, "workflow.yaml")),
					httpmock.StringResponse("{}"))

				reg.Register(
					httpmock.REST("GET", fmt.Sprintf("repos/%s/actions/workflows/workflow.yaml", repo)),
					httpmock.StringResponse(getWorkflowResponse))

				v := url.Values{}
				v.Set("per_page", "100")

				reg.Register(
					httpmock.QueryMatcher("GET", fmt.Sprintf("repos/%s/actions/workflows/456/runs", repo), v),
					httpmock.StringResponse(getWorkflowRunsResponse))

				q := url.Values{}
				q.Set("per_page", "100")
				q.Set("page", "1")

				// TODO: is this correct? is it the correct response?
				reg.Register(
					httpmock.QueryMatcher("GET", fmt.Sprintf("repos/%s/actions/workflows", repo), q),
					httpmock.StringResponse(getWorkflowRunsResponse))

				// TODO: is this correct? is it the correct response?
				reg.Register(
					httpmock.REST("GET", fmt.Sprintf("repos/%s/actions/workflows/456", repo)),
					httpmock.StringResponse(getWorkflowResponse))

				reg.Register(
					httpmock.REST("GET", fmt.Sprintf("repos/%s/actions/runs/123", repo)),
					httpmock.StringResponse(`{
						"id": 123,
						"workflow_id": 456
					}`))

				reg.Register(
					httpmock.REST("GET", fmt.Sprintf("repos/%s/actions/workflows/456", repo)),
					httpmock.StringResponse(getWorkflowResponse))

				reg.Register(
					httpmock.REST("GET", fmt.Sprintf("repos/%s/actions/workflows/456", repo)),
					httpmock.StringResponse(getWorkflowResponse))

				reg.Register(
					httpmock.REST("GET", fmt.Sprintf("repos/%s/actions/runs/123", repo)),
					httpmock.StringResponse(`{
						"id": 123,
						"workflow_id": 456,
						"status": "completed",
						"conclusion": "success"
					}`))

				reg.Register(
					httpmock.REST("GET", fmt.Sprintf("repos/%s/actions/runs/123/attempts/1/jobs", repo)),
					httpmock.StringResponse(getJobsResponse))

				reg.Register(
					httpmock.REST("GET", fmt.Sprintf("repos/%s/check-runs/123/annotations", repo)),
					httpmock.StringResponse("[]"))
			},
			wantOut: "Refreshing run status every 2 seconds. Press Ctrl+C to quit.\n\nhttps://github.com/OWNER/REPO/actions/runs/123\n\n\nJOBS\n✓ build in 1m59s (ID 123)\n  ✓ Run actions/checkout@v2\n  ✓ Test\n",
		}, {
			name: "unsuccessful workflow run",
			opts: &workflowDispatchOptions{
				inputs:   `{"foo": "bar"}`,
				workflow: workflow,
			},
			httpStubs: func(reg *httpmock.Registry) {
				// TODO: test that proper request body is sent on POSTs
				reg.Register(
					httpmock.REST("POST", fmt.Sprintf("repos/%s/actions/workflows/%s/dispatches", repo, "workflow.yaml")),
					httpmock.StringResponse("{}"))

				reg.Register(
					httpmock.REST("GET", fmt.Sprintf("repos/%s/actions/workflows/workflow.yaml", repo)),
					httpmock.StringResponse(getWorkflowResponse))

				v := url.Values{}
				v.Set("per_page", "100")

				reg.Register(
					httpmock.QueryMatcher("GET", fmt.Sprintf("repos/%s/actions/workflows/456/runs", repo), v),
					httpmock.StringResponse(getWorkflowRunsResponse))

				q := url.Values{}
				q.Set("per_page", "100")
				q.Set("page", "1")

				// TODO: is this correct? is it the correct response?
				reg.Register(
					httpmock.QueryMatcher("GET", fmt.Sprintf("repos/%s/actions/workflows", repo), q),
					httpmock.StringResponse(getWorkflowRunsResponse))

				// TODO: is this correct? is it the correct response?
				reg.Register(
					httpmock.REST("GET", fmt.Sprintf("repos/%s/actions/workflows/456", repo)),
					httpmock.StringResponse(getWorkflowResponse))

				// TODO: is this correct? is it the correct response?
				reg.Register(
					httpmock.REST("GET", fmt.Sprintf("repos/%s/actions/workflows/456", repo)),
					httpmock.StringResponse(getWorkflowResponse))

				reg.Register(
					httpmock.REST("GET", fmt.Sprintf("repos/%s/actions/runs/123", repo)),
					httpmock.StringResponse(`{
						"id": 123,
						"workflow_id": 456
					}`))

				reg.Register(
					httpmock.REST("GET", fmt.Sprintf("repos/%s/actions/workflows/456", repo)),
					httpmock.StringResponse(getWorkflowResponse))

				reg.Register(
					httpmock.REST("GET", fmt.Sprintf("repos/%s/actions/runs/123", repo)),
					httpmock.StringResponse(`{
						"id": 123,
						"workflow_id": 456,
						"status": "completed",
						"conclusion": "failure"
					}`))

				reg.Register(
					httpmock.REST("GET", fmt.Sprintf("repos/%s/actions/runs/123/attempts/1/jobs", repo)),
					httpmock.StringResponse(getFailingJobsResponse))

				reg.Register(
					httpmock.REST("GET", fmt.Sprintf("repos/%s/check-runs/123/annotations", repo)),
					httpmock.StringResponse("[]"))
			},
			wantOut: "Refreshing run status every 2 seconds. Press Ctrl+C to quit.\n\nhttps://github.com/OWNER/REPO/actions/runs/123\n\n\nJOBS\n✓ build in 1m59s (ID 123)\n  ✓ Run actions/checkout@v2\n  X Test\n",
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
