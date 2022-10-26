package cmd

import (
	"testing"

	"github.com/cli/cli/pkg/httpmock"
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
			name: "basic",
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
		}}

	for _, tt := range tests {
		ios, _, stdout, _ := iostreams.Test()
		ios.SetStdoutTTY(false)
		ios.SetAlternateScreenBufferEnabled(false)
		tt.opts.IO = ios

		reg := &httpmock.Registry{}
		tt.httpStubs(reg)

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
