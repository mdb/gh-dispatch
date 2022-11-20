package dispatch

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGHRepo(t *testing.T) {
	tests := []struct {
		name         string
		wantName     string
		wantOwner    string
		wantHost     string
		wantFullName string
		wantErr      bool
		errMsg       string
	}{
		{
			name:         "foo/bar",
			wantOwner:    "foo",
			wantName:     "bar",
			wantFullName: "foo/bar",
			wantHost:     "github.com",
			wantErr:      false,
		}, {
			name:         "other-github.com/foo/bar",
			wantOwner:    "foo",
			wantName:     "bar",
			wantFullName: "foo/bar",
			wantHost:     "other-github.com",
			wantErr:      false,
		}, {
			name:    "bar",
			wantErr: true,
			errMsg:  "invalid repository name",
		}, {
			name:    "",
			wantErr: true,
			errMsg:  "invalid repository name",
		}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ghr, err := newGHRepo(tt.name)

			if tt.wantErr {
				assert.EqualError(t, err, tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantFullName, ghr.RepoFullName())
				assert.Equal(t, tt.wantOwner, ghr.RepoOwner())
				assert.Equal(t, tt.wantName, ghr.RepoName())
				assert.Equal(t, tt.wantHost, ghr.RepoHost())
			}
		})
	}
}
