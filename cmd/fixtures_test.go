package cmd

var (
	getWorkflowsResponse string = `{
		"total_count": 1,
		"workflows": [{
			"id": 456,
			"name": "foo"
		}]
	}`

	getWorkflowResponse string = `{
		"id": 456
	}`

	getWorkflowRunsResponse string = `{
		"total_count": 1,
		"workflow_runs": [{
			"id": 123,
			"workflow_id": 456,
			"name": "foo",
			"status": "queued",
			"conclusion": null
		}]
	}`

	getJobsResponse string = `{
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
	}`

	getFailingJobsResponse string = `{
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
					"conclusion": "failure",
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
	}`
)
