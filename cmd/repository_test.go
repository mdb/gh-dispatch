package cmd

import (
	"net/url"
	"testing"

	"github.com/cli/cli/v2/pkg/httpmock"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/stretchr/testify/assert"
)

func TestRepositoryDispatchRun(t *testing.T) {
	tests := []struct {
		name      string
		opts      *repositoryDispatchOptions
		httpStubs func(*httpmock.Registry)
		wantErr   bool
		errMsg    string
		wantOut   string
	}{
		{
			name: "no specified workflow",
			opts: &repositoryDispatchOptions{
				Repo:      "OWNER/REPO",
				EventType: "hello",
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("POST", "repos/OWNER/REPO/dispatches"),
					httpmock.StringResponse("{}"))
			},
			wantOut: "",
		}, {
			name: "specified workflow",
			opts: &repositoryDispatchOptions{
				Repo:      "OWNER/REPO",
				EventType: "hello",
				Workflow:  "foo",
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("POST", "repos/OWNER/REPO/dispatches"),
					httpmock.StringResponse("{}"))

				v := url.Values{}
				v.Set("name", "foo")
				v.Set("event", "repository_dispatch")

				reg.Register(
					httpmock.QueryMatcher("GET", "repos/OWNER/REPO/actions/runs", v),
					httpmock.StringResponse(`{
						"total_count": 1,
						"workflow_runs": [{
							"id": 123,
							"workflow_id": 456,
							"name": "foo",
							"status": "queued",
							"conclusion": null
						}]
					}`))

				reg.Register(
					httpmock.REST("GET", "repos/OWNER/REPO/actions/runs/123"),
					httpmock.StringResponse(`{
						"id": 123
					}`))

				reg.Register(
					httpmock.REST("GET", "repos/OWNER/REPO/actions/runs/123"),
					httpmock.StringResponse(`{
						"id": 123
					}`))

				reg.Register(
					httpmock.REST("GET", "repos/OWNER/REPO/actions/runs/123"),
					httpmock.StringResponse(`{
						"id": 123,
						"status": "completed"
					}`))

				reg.Register(
					httpmock.REST("GET", "repos/OWNER/REPO/actions/runs/123/attempts/1/jobs"),
					httpmock.StringResponse(`{
  					"total_count": 1,
						"jobs": [{
							"id": 123,
							"run_id": 29679449,
							"run_url": "https://api.github.com/repos/octo-org/octo-repo/actions/runs/29679449",
							"node_id": "MDEyOldvcmtmbG93IEpvYjM5OTQ0NDQ5Ng==",
							"head_sha": "f83a356604ae3c5d03e1b46ef4d1ca77d64a90b0",
							"url": "https://api.github.com/repos/octo-org/octo-repo/actions/jobs/399444496",
							"html_url": "https://github.com/octo-org/octo-repo/runs/399444496",
							"status": "completed",
							"conclusion": "success",
							"started_at": "2020-01-20T17:42:40Z",
							"completed_at": "2020-01-20T17:44:39Z",
							"name": "build",
							"steps": [
								{
									"name": "Run actions/checkout@v2",
									"status": "completed",
									"conclusion": "success",
									"number": 2,
									"started_at": "2020-01-20T09:42:41.000-08:00",
									"completed_at": "2020-01-20T09:42:45.000-08:00"
								},
								{
									"name": "Test",
									"status": "completed",
									"conclusion": "success",
									"number": 3,
									"started_at": "2020-01-20T09:42:45.000-08:00",
									"completed_at": "2020-01-20T09:42:45.000-08:00"
								}
							],
							"check_run_url": "https://api.github.com/repos/octo-org/octo-repo/check-runs/399444496",
							"labels": [
								"self-hosted",
								"foo",
								"bar"
							],
							"runner_id": 1,
							"runner_name": "my runner",
							"runner_group_id": 2,
							"runner_group_name": "my runner group"
						}]
					}`))

				// TODO: abstract response body to common location
				reg.Register(
					httpmock.REST("GET", "repos/OWNER/REPO/actions/runs/123/attempts/1/jobs"),
					httpmock.StringResponse(`{
  					"total_count": 1,
						"jobs": [{
							"id": 123,
							"run_id": 29679449,
							"run_url": "https://api.github.com/repos/octo-org/octo-repo/actions/runs/29679449",
							"node_id": "MDEyOldvcmtmbG93IEpvYjM5OTQ0NDQ5Ng==",
							"head_sha": "f83a356604ae3c5d03e1b46ef4d1ca77d64a90b0",
							"url": "https://api.github.com/repos/octo-org/octo-repo/actions/jobs/399444496",
							"html_url": "https://github.com/octo-org/octo-repo/runs/399444496",
							"status": "completed",
							"conclusion": "success",
							"started_at": "2020-01-20T17:42:40Z",
							"completed_at": "2020-01-20T17:44:39Z",
							"name": "build",
							"steps": [
								{
									"name": "Run actions/checkout@v2",
									"status": "completed",
									"conclusion": "success",
									"number": 2,
									"started_at": "2020-01-20T09:42:41.000-08:00",
									"completed_at": "2020-01-20T09:42:45.000-08:00"
								},
								{
									"name": "Test",
									"status": "completed",
									"conclusion": "success",
									"number": 3,
									"started_at": "2020-01-20T09:42:45.000-08:00",
									"completed_at": "2020-01-20T09:42:45.000-08:00"
								}
							],
							"check_run_url": "https://api.github.com/repos/octo-org/octo-repo/check-runs/399444496",
							"labels": [
								"self-hosted",
								"foo",
								"bar"
							],
							"runner_id": 1,
							"runner_name": "my runner",
							"runner_group_id": 2,
							"runner_group_name": "my runner group"
						}]
					}`))

				reg.Register(
					httpmock.REST("GET", "repos/OWNER/REPO/check-runs/123/annotations"),
					httpmock.StringResponse("[]"))
			},
			wantOut: "Refreshing run status every 3 seconds. Press Ctrl+C to quit.\n\nhttps://github.com/OWNER/REPO/actions/runs/123\n\n\nJOBS\n✓ build in 1m59s (ID 123)\n  ✓ Run actions/checkout@v2\n  ✓ Test\n",
		}, {
			name: "malformed JSON response",
			opts: &repositoryDispatchOptions{
				Repo:      "OWNER/REPO",
				EventType: "hello",
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("POST", "repos/OWNER/REPO/dispatches"),
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

		tt.opts.IO = ios
		tt.opts.HTTPTransport = reg
		tt.opts.AuthToken = "123"

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
