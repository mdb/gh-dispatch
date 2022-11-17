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

func TestRepositoryDispatchRun(t *testing.T) {
	ghRepo := &ghRepo{
		Name:  "REPO",
		Owner: "OWNER",
	}
	repo := ghRepo.RepoFullName()
	event := "repository_dispatch"

	tests := []struct {
		name      string
		opts      *repositoryDispatchOptions
		httpStubs func(*httpmock.Registry)
		wantErr   bool
		errMsg    string
		wantOut   string
	}{
		{
			name: "successful workflow run",
			opts: &repositoryDispatchOptions{
				eventType: "hello",
				workflow:  "foo",
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("POST", fmt.Sprintf("repos/%s/dispatches", repo)),
					httpmock.StringResponse("{}"))

				reg.Register(
					httpmock.REST("GET", fmt.Sprintf("repos/%s/actions/workflows", repo)),
					httpmock.StringResponse(getWorkflowsResponse))

				reg.Register(
					httpmock.GraphQL("query UserCurrent{viewer{login}}"),
					httpmock.StringResponse(currentUserResponse))

				v := url.Values{}
				v.Set("per_page", "50")

				reg.Register(
					httpmock.QueryMatcher("GET", fmt.Sprintf("repos/%s/actions/workflows/456/runs", repo), v),
					httpmock.StringResponse(fmt.Sprintf(getWorkflowRunsResponse, event, repo)))

				q := url.Values{}
				q.Set("per_page", "100")
				q.Set("page", "1")

				// TODO: is this correct? is it the correct response?
				reg.Register(
					httpmock.QueryMatcher("GET", fmt.Sprintf("repos/%s/actions/workflows", repo), q),
					httpmock.StringResponse(fmt.Sprintf(getWorkflowRunsResponse, event, repo)))

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
						"workflow_id": 456,
						"event": "repository_dispatch"
					}`))

				reg.Register(
					httpmock.REST("GET", fmt.Sprintf("repos/%s/actions/workflows/456", repo)),
					httpmock.StringResponse(getWorkflowResponse))

				reg.Register(
					httpmock.REST("GET", fmt.Sprintf("repos/%s/actions/runs/123", repo)),
					httpmock.StringResponse(fmt.Sprintf(`{
						"id": 123,
						"workflow_id": 456,
						"status": "completed",
						"event": "repository_dispatch",
						"conclusion": "success",
						"jobs_url": "https://api.github.com/repos/%s/actions/runs/123/jobs"
					}`, repo)))

				reg.Register(
					httpmock.REST("GET", fmt.Sprintf("repos/%s/actions/runs/123/jobs", repo)),
					httpmock.StringResponse(getJobsResponse))

				reg.Register(
					httpmock.REST("GET", fmt.Sprintf("repos/%s/check-runs/123/annotations", repo)),
					httpmock.StringResponse("[]"))
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
			httpStubs: func(reg *httpmock.Registry) {
				// TODO: test that proper request body is sent on POSTs
				reg.Register(
					httpmock.REST("POST", fmt.Sprintf("repos/%s/dispatches", repo)),
					httpmock.StringResponse("{}"))

				reg.Register(
					httpmock.REST("GET", fmt.Sprintf("repos/%s/actions/workflows", repo)),
					httpmock.StringResponse(getWorkflowsResponse))

				reg.Register(
					httpmock.GraphQL("query UserCurrent{viewer{login}}"),
					httpmock.StringResponse(currentUserResponse))

				v := url.Values{}
				v.Set("per_page", "50")

				reg.Register(
					httpmock.QueryMatcher("GET", fmt.Sprintf("repos/%s/actions/workflows/456/runs", repo), v),
					httpmock.StringResponse(fmt.Sprintf(getWorkflowRunsResponse, event, repo)))

				q := url.Values{}
				q.Set("per_page", "100")
				q.Set("page", "1")

				// TODO: is this correct? is it the correct response?
				reg.Register(
					httpmock.QueryMatcher("GET", fmt.Sprintf("repos/%s/actions/workflows", repo), q),
					httpmock.StringResponse(fmt.Sprintf(getWorkflowRunsResponse, event, repo)))

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
						"workflow_id": 456,
						"event": "repository_dispatch"
					}`))

				reg.Register(
					httpmock.REST("GET", fmt.Sprintf("repos/%s/actions/workflows/456", repo)),
					httpmock.StringResponse(getWorkflowResponse))

				reg.Register(
					httpmock.REST("GET", fmt.Sprintf("repos/%s/actions/runs/123", repo)),
					httpmock.StringResponse(fmt.Sprintf(`{
						"id": 123,
						"workflow_id": 456,
						"status": "completed",
						"event": "repository_dispatch",
						"conclusion": "failure",
						"jobs_url": "https://api.github.com/repos/%s/actions/runs/123/jobs"
					}`, repo)))

				reg.Register(
					httpmock.REST("GET", fmt.Sprintf("repos/%s/actions/runs/123/jobs", repo)),
					httpmock.StringResponse(getFailingJobsResponse))

				reg.Register(
					httpmock.REST("GET", fmt.Sprintf("repos/%s/check-runs/123/annotations", repo)),
					httpmock.StringResponse("[]"))
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
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("POST", fmt.Sprintf("repos/%s/dispatches", repo)),
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
			err := repositoryDispatchRun(tt.opts)

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
