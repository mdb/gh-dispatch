package cmd

import (
	"testing"

	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/stretchr/testify/assert"
)

func TestRepositoryDispatchRun(t *testing.T) {
	tests := []struct {
		name    string
		opts    *repositoryDispatchOptions
		wantErr bool
		errMsg  string
		wantOut string
	}{
		{
			name: "repo does not exist",
			opts: &repositoryDispatchOptions{
				Repo:      "owner/repo",
				EventType: "hello",
			},
			wantErr: true,
			errMsg:  "HTTP 404: Not Found (https://api.github.com/repos/owner/repo/dispatches)",
			wantOut: "",
		}}

	for _, tt := range tests {
		ios, _, stdout, _ := iostreams.Test()
		ios.SetStdoutTTY(false)
		ios.SetAlternateScreenBufferEnabled(false)
		tt.opts.IO = ios

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
		})
	}
}
