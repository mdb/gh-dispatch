package dispatch

import (
	"errors"
	"fmt"
	"strings"

	"github.com/cli/go-gh/pkg/auth"
	"github.com/spf13/cobra"
)

func getRepoOption(cmd *cobra.Command) (*ghRepo, error) {
	r, _ := cmd.Flags().GetString("repo")
	if r == "" {
		return nil, errors.New("a --repo must be specified in the [HOST/]OWNER/REPO format")
	}

	repo, err := newGHRepo(r)
	if err != nil {
		return nil, err
	}

	return repo, nil
}

// ghRepo satisfies the ghrepo interface.
// In the context of gh-dispatch, it enables the reuse of
// functions packaged in the upstream github.com/cli/cli
// codebase for rendering GH Actions run output.
// See github.com/cli/cli/v2/internal/ghrepo.
type ghRepo struct {
	Owner string
	Name  string
	Host  string
}

func newGHRepo(name string) (*ghRepo, error) {
	defaultHost, _ := auth.DefaultHost()
	nameParts := strings.Split(name, "/")

	switch len(nameParts) {
	case 2:
		return &ghRepo{
			Owner: nameParts[0],
			Name:  nameParts[1],
			Host:  defaultHost,
		}, nil
	case 3:
		return &ghRepo{
			Owner: nameParts[1],
			Name:  nameParts[2],
			Host:  nameParts[0],
		}, nil
	default:
		return nil, errors.New("invalid repository name")
	}
}

func (r ghRepo) RepoName() string {
	return r.Name
}

func (r ghRepo) RepoOwner() string {
	return r.Owner
}

func (r ghRepo) RepoHost() string {
	return r.Host
}

func (r ghRepo) RepoFullName() string {
	return fmt.Sprintf("%s/%s", r.RepoOwner(), r.RepoName())
}
