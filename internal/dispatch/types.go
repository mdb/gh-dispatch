package dispatch

import (
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/cli/go-gh/v2/pkg/api"
)

type dispatchOptions struct {
	repo   *ghRepo
	client *api.RESTClient
	io     *iostreams.IOStreams
}
