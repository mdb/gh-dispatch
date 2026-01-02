package dispatch

import (
	"net/http"

	"github.com/cli/cli/v2/pkg/iostreams"
)

type dispatchOptions struct {
	repo       *ghRepo
	httpClient *http.Client
	io         *iostreams.IOStreams
}
