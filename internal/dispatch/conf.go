package dispatch

import (
	"github.com/cli/go-gh/pkg/auth"
	"github.com/cli/go-gh/pkg/config"
)

// Conf implements the cliapi tokenGetter interface.
// In the context of gh-dispatch, its only purpose is to provide
// a configuration to the GH HTTP client configuration that
// enables the retrieval of an auth token in the same standard way
// that gh itself retrieves the auth token.
type Conf struct {
	*config.Config
}

// AuthToken implements the cliapi tokenGetter interface
// by providing a method for retrieving the auth token.
func (c *Conf) AuthToken(hostname string) (string, string) {
	return auth.TokenForHost(hostname)
}
