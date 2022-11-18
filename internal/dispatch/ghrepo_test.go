package dispatch

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGHRepo(t *testing.T) {
	ghRepo := &ghRepo{
		Name:  "REPO",
		Owner: "OWNER",
	}

	assert.Equal(t, "OWNER/REPO", ghRepo.RepoFullName())
	assert.Equal(t, "OWNER", ghRepo.RepoOwner())
	assert.Equal(t, "REPO", ghRepo.RepoName())
	assert.Equal(t, "github.com", ghRepo.RepoHost())
}
