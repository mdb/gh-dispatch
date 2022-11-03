package main

import (
	"fmt"
	"os"
	"os/exec"
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

func TestCommand(t *testing.T) {
	basicOut := "gh dispatch: Trigger a GitHub dispatch event and watch the resulting GitHub Actions run\n\nUsage:\n  gh [command]\n\nExamples:\nTODO\n\nAvailable Commands:\n  completion  Generate the autocompletion script for the specified shell\n  help        Help about any command\n  repository  The 'repository' subcommand triggers repository dispatch events\n  workflow    The 'workflow' subcommand triggers workflow dispatch events\n\nFlags:\n  -h, --help          help for gh\n  -R, --repo string   The targeted repository's full name (in 'owner/repo' format)\n  -v, --version       version for gh\n\nUse \"gh [command] --help\" for more information about a command.\n"
	tests := []struct {
		arg     string
		wantOut string
		errMsg  string
		wantErr bool
	}{{
		arg:     "",
		wantOut: basicOut,
	}, {
		arg:     "--help",
		wantOut: basicOut,
	}, {
		arg:     "help",
		wantOut: basicOut,
	}}

	for _, test := range tests {
		t.Run(fmt.Sprintf("when 'gh dispatch' is passed '%s'", test.arg), func(t *testing.T) {
			output, err := exec.Command("./gh-dispatch", test.arg).CombinedOutput()

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
