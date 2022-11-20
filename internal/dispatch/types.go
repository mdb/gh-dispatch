package dispatch

import (
	"errors"
	"net/http"

	"github.com/cli/cli/v2/pkg/cmd/run/shared"
	"github.com/cli/cli/v2/pkg/iostreams"
)

var errUnspecifiedRepo = errors.New("A --repo must be specified in the [HOST/]OWNER/REPO format")

type workflowRun struct {
	ID         int64         `json:"id"`
	WorkflowID int           `json:"workflow_id"`
	Name       string        `json:"name"`
	Status     shared.Status `json:"status"`
	Conclusion string        `json:"conclusion"`
}

type workflowRunsResponse struct {
	WorkflowRuns []workflowRun `json:"workflow_runs"`
}

type dispatchOptions struct {
	repo       *ghRepo
	httpClient *http.Client
	io         *iostreams.IOStreams
}
