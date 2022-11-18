package dispatch

import (
	"fmt"

	"github.com/cli/go-gh/pkg/auth"
)

// ghRepo satisfies the ghrepo interface.
// In the context of gh-dispatch, it enables the reuse of
// functions packaged in the upstream github.com/cli/cli
// codebase for rendering GH Actions run output.
// See github.com/cli/cli/v2/internal/ghrepo.
type ghRepo struct {
	Name  string
	Owner string
}

func (r ghRepo) RepoName() string {
	return r.Name
}

func (r ghRepo) RepoOwner() string {
	return r.Owner
}

func (r ghRepo) RepoHost() string {
	host, _ := auth.DefaultHost()

	return host
}

func (r ghRepo) RepoFullName() string {
	return fmt.Sprintf("%s/%s", r.RepoOwner(), r.RepoName())
}
