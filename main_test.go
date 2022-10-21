package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
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

func TestRootCommand(t *testing.T) {
	tests := []struct {
		arg     string
		outputs []string
		err     error
	}{{
		arg:     "",
		outputs: []string{"gh dispatch: Trigger a GitHub dispatch event and watch the resulting GitHub Actions run"},
		err:     nil,
	}}

	for _, test := range tests {
		t.Run(fmt.Sprintf("when 'gh dispatch' is passed '%s'", test.arg), func(t *testing.T) {
			output, err := exec.Command("./gh-dispatch", test.arg).CombinedOutput()

			if test.err == nil && err != nil {
				t.Errorf("expected '%s' not to error; got '%v'", test.arg, err)
			}

			if test.err != nil && err == nil {
				t.Errorf("expected '%s' to error with '%s', but it didn't error", test.arg, test.err.Error())
			}

			if test.err != nil && err != nil && test.err.Error() != err.Error() {
				t.Errorf("expected '%s' to error with '%s'; got '%s'", test.arg, test.err.Error(), err.Error())
			}

			for _, o := range test.outputs {
				if !strings.Contains(string(output), o) {
					t.Errorf("expected '%s' to include output '%s'; got '%s'", test.arg, o, output)
				}
			}
		})
	}
}
