package dispatch

import (
	"net/http"

	"github.com/cli/cli/v2/pkg/iostreams"
	cliapi "github.com/cli/go-gh/v2/pkg/api"
)

type dispatchOptions struct {
	repo       *ghRepo
	client     *cliapi.RESTClient
	httpClient *http.Client
	io         *iostreams.IOStreams
}
