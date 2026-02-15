//go:build acceptance
// +build acceptance

package main

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/stretchr/testify/assert"
)

func TestRootAcceptance(t *testing.T) {
	basicOut := heredoc.Doc(`Send a workflow_dispatch or repository_dispatch event and watch the resulting
GitHub Actions run.

Usage:
  gh [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  repository  Send a repository dispatch event and watch the resulting GitHub Actions run
  workflow    Send a workflow dispatch event and watch the resulting GitHub Actions run

Flags:
  -h, --help          help for gh
  -R, --repo string   The targeted repository's full name (default "github.com/mdb/gh-dispatch")
  -v, --version       version for gh

Use "gh [command] --help" for more information about a command.
`)
	tests := []struct {
		args    []string
		wantOut string
		errMsg  string
		wantErr bool
	}{{
		args:    []string{"dispatch"},
		wantOut: basicOut,
	}, {
		args: []string{
			"dispatch",
			"--help",
		},
		wantOut: basicOut,
	}, {
		args: []string{
			"dispatch",
			"help",
		},
		wantOut: basicOut,
	}}

	for _, test := range tests {
		t.Run(fmt.Sprintf("when passed '%s'", strings.Join(test.args, " ")), func(t *testing.T) {
			output, err := exec.Command("gh", test.args...).CombinedOutput()

			if test.wantErr {
				assert.EqualError(t, err, test.errMsg)
			} else {
				assert.NoError(t, err)
			}

			if got := string(output); got != test.wantOut {
				t.Errorf("got stdout:\n%q\nwant:\n%q", got, test.wantOut)
			}
		})
	}
}

func TestRepositoryAcceptance(t *testing.T) {
	tests := []struct {
		args    []string
		wantOut []string
		errMsg  string
		wantErr bool
	}{{
		args: []string{
			"dispatch",
			"repository",
			"--event-type=hello",
			`--client-payload={"name": "Mike"}`,
			"--workflow=Hello",
		},
		wantOut: []string{
			"Refreshing run status every 2 seconds. Press Ctrl+C to quit.\n\nhttps://github.com/mdb/gh-dispatch/actions/runs",
			"JOBS\n* hello (ID",
			")\n  ✓ Set up job",
			"\n  ✓ say-hello",
			"\n  ✓ Complete job\n",
		},
	}, {
		args: []string{
			"dispatch",
			"repository",
			"--repo=mdb/gh-dispatch",
			"--event-type=hello",
			`--client-payload={"force_fail": true}`,
			"--workflow=Hello",
		},
		wantOut: []string{
			"Refreshing run status every 2 seconds. Press Ctrl+C to quit.\n\nhttps://github.com/mdb/gh-dispatch/actions/runs",
			"JOBS\n* hello (ID",
			")\n  ✓ Set up job",
			"\n  X say-hello",
			"\n  ✓ Complete job\n",
		},
		wantErr: true,
		errMsg:  "exit status 1",
	}}

	for _, test := range tests {
		t.Run(fmt.Sprintf("when passed '%s'", strings.Join(test.args, " ")), func(t *testing.T) {
			output, err := exec.Command("gh", test.args...).CombinedOutput()

			if test.wantErr {
				assert.EqualError(t, err, test.errMsg)
			} else {
				assert.NoError(t, err)
			}

			got := string(output)
			for _, out := range test.wantOut {
				if !strings.Contains(got, out) {
					t.Errorf("expected stdout to include:\n%q\ngot:\n%q", out, got)
				}
			}

			// sleep for 2s so that each test fetches the correct run ID
			// Until https://github.com/mdb/gh-dispatch/issues/9 is addressed, `gh dispatch`
			// does not reliably find the correct workflow run correctly corresponding to
			// the dispatch event triggered by `gh dispatch`.
			duration, _ := time.ParseDuration("2s")
			time.Sleep(duration)
		})
	}
}

func TestWorkflowAcceptance(t *testing.T) {
	tests := []struct {
		args    []string
		wantOut []string
		errMsg  string
		wantErr bool
	}{{
		args: []string{
			"dispatch",
			"workflow",
			`--inputs={"name": "Mike", "force_fail": "false"}`,
			"--workflow=workflow_dispatch.yaml",
		},
		wantOut: []string{
			"Refreshing run status every 2 seconds. Press Ctrl+C to quit.\n\nhttps://github.com/mdb/gh-dispatch/actions/runs",
			"JOBS\n* goodbye (ID",
			")\n  ✓ Set up job",
			"\n  ✓ say-goodbye",
			"\n  ✓ Complete job\n",
		},
	}, {
		args: []string{
			"dispatch",
			"workflow",
			`--inputs={"name": "Mike", "force_fail": "true"}`,
			"--workflow=workflow_dispatch.yaml",
		},
		wantOut: []string{
			"Refreshing run status every 2 seconds. Press Ctrl+C to quit.\n\nhttps://github.com/mdb/gh-dispatch/actions/runs",
			"JOBS\n* goodbye (ID",
			")\n  ✓ Set up job",
			"\n  X say-goodbye",
			"\n  ✓ Complete job\n",
		},
		wantErr: true,
		errMsg:  "exit status 1",
	}}

	for _, test := range tests {
		t.Run(fmt.Sprintf("when passed '%s'", strings.Join(test.args, " ")), func(t *testing.T) {
			output, err := exec.Command("gh", test.args...).CombinedOutput()

			if test.wantErr {
				assert.EqualError(t, err, test.errMsg)
			} else {
				assert.NoError(t, err)
			}

			got := string(output)
			for _, out := range test.wantOut {
				if !strings.Contains(got, out) {
					t.Errorf("expected stdout to include:\n%q\ngot:\n%q", out, got)
				}
			}

			// sleep for 2s so that each test fetches the correct run ID
			// Until https://github.com/mdb/gh-dispatch/issues/9 is addressed, `gh dispatch`
			// does not reliably find the correct workflow run correctly corresponding to
			// the dispatch event triggered by `gh dispatch`.
			duration, _ := time.ParseDuration("2s")
			time.Sleep(duration)
		})
	}
}
