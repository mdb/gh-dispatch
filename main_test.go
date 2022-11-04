// +build acceptance

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	// compile a 'gh-dispatch' for for use in running tests
	exe := exec.Command("go", "build", "-ldflags", "-X main.version=test", "-o", "gh-dispatch")
	err := exe.Run()
	if err != nil {
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func TestRootAcceptance(t *testing.T) {
	basicOut := "gh dispatch: Trigger a GitHub dispatch event and watch the resulting GitHub Actions run\n\nUsage:\n  gh [command]\n\nExamples:\nTODO\n\nAvailable Commands:\n  completion  Generate the autocompletion script for the specified shell\n  help        Help about any command\n  repository  The 'repository' subcommand triggers repository dispatch events\n  workflow    The 'workflow' subcommand triggers workflow dispatch events\n\nFlags:\n  -h, --help          help for gh\n  -R, --repo string   The targeted repository's full name (in 'owner/repo' format)\n  -v, --version       version for gh\n\nUse \"gh [command] --help\" for more information about a command.\n"
	tests := []struct {
		args    []string
		wantOut string
		errMsg  string
		wantErr bool
	}{{
		args:    []string{""},
		wantOut: basicOut,
	}, {
		args:    []string{"--help"},
		wantOut: basicOut,
	}, {
		args:    []string{"help"},
		wantOut: basicOut,
	}}

	for _, test := range tests {
		t.Run(fmt.Sprintf("when 'gh dispatch' is passed '%s'", strings.Join(test.args, " ")), func(t *testing.T) {
			output, err := exec.Command("./gh-dispatch", test.args...).CombinedOutput()

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
			"repository",
			"--repo=mdb/gh-dispatch",
			"--event-type=hello",
			`--client-payload={"name": "Mike"}`,
			"--workflow=Hello",
		},
		wantOut: []string{
			"Refreshing run status every 2 seconds. Press Ctrl+C to quit.\n\nhttps://github.com/mdb/gh-dispatch/actions/runs",
			"JOBS\n* hello (ID",
			")\n  ✓ Set up job",
			"\n  ✓ Run actions/checkout@v3",
			"\n  ✓ say-hello",
			"\n  ✓ Post Run actions/checkout@v3\n",
			"\n  ✓ Complete job\n",
		},
	}, {
		args: []string{
			"repository",
			"--repo=mdb/gh-dispatch",
			"--event-type=hello",
			`--client-payload={"force_fail": true}`,
			"--workflow=Hello",
		},
		wantOut: []string{
			"TODO",
		},
	}}

	for _, test := range tests {
		t.Run(fmt.Sprintf("when 'gh dispatch' is passed '%s'", strings.Join(test.args, " ")), func(t *testing.T) {
			output, err := exec.Command("./gh-dispatch", test.args...).CombinedOutput()

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
		})
	}
}
