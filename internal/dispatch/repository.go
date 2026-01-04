package dispatch

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/MakeNowJust/heredoc"
	cliapi "github.com/cli/cli/v2/api"
	runShared "github.com/cli/cli/v2/pkg/cmd/run/shared"
	"github.com/cli/cli/v2/pkg/cmd/workflow/shared"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/spf13/cobra"
)

type repositoryDispatchRequest struct {
	EventType     string `json:"event_type"`
	ClientPayload any    `json:"client_payload"`
}

type repositoryDispatchOptions struct {
	clientPayload any
	eventType     string
	workflow      string
	dispatchOptions
}

// NewCmdRepository returns a new repository command.
func NewCmdRepository() *cobra.Command {
	var (
		repositoryEventType     string
		repositoryClientPayload string
		repositoryWorkflow      string
	)

	cmd := &cobra.Command{
		Use: heredoc.Doc(`
		repository \
			--repo [owner/repo] \
			--event-type [event-type] \
			--client-payload [json-string] \
			--workflow [workflow-name]
	`),
		Short: "Send a repository dispatch event and watch the resulting GitHub Actions run",
		Long: heredoc.Doc(`
		This command sends a repository dispatch event and attempts to find and watch the
		resulting GitHub Actions run whose name is specified as '--workflow'.

		Note that the command assumes the specified workflow supports a repository_dispatch
		'on' trigger. Also note that the command is vulnerable to race conditions and may
		watch an unrelated GitHub Actions workflow run in the event that multiple runs of
		the specified workflow are running concurrently.
	`),
		Example: heredoc.Doc(`
		gh dispatch repository \
			--repo mdb/gh-dispatch \
			--event-type 'hello' \
			--client-payload '{"name": "Mike"}' \
			--workflow Hello
	`),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, err := getRepoOption(cmd)
			if err != nil {
				return err
			}

			b := []byte(repositoryClientPayload)
			var repoClientPayload any
			json.Unmarshal(b, &repoClientPayload)

			ios := iostreams.System()
			ghClient, _ := cliapi.NewHTTPClient(cliapi.HTTPClientOptions{})
			dOptions := dispatchOptions{
				repo:       repo,
				httpClient: ghClient,
				io:         ios,
			}

			return repositoryDispatchRun(&repositoryDispatchOptions{
				clientPayload:   repoClientPayload,
				eventType:       repositoryEventType,
				workflow:        repositoryWorkflow,
				dispatchOptions: dOptions,
			})
		},
	}

	cmd.Flags().StringVarP(&repositoryEventType, "event-type", "e", "", "The repository dispatch event type.")
	cmd.MarkFlagRequired("event-type")
	cmd.Flags().StringVarP(&repositoryClientPayload, "client-payload", "p", "", "The repository dispatch event client payload JSON string.")
	cmd.MarkFlagRequired("client-payload")
	cmd.Flags().StringVarP(&repositoryWorkflow, "workflow", "w", "", "The resulting GitHub Actions workflow name.")
	cmd.MarkFlagRequired("workflow")

	return cmd
}

func repositoryDispatchRun(opts *repositoryDispatchOptions) error {
	ghClient := cliapi.NewClientFromHTTP(opts.httpClient)

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(repositoryDispatchRequest{
		EventType:     opts.eventType,
		ClientPayload: opts.clientPayload,
	})
	if err != nil {
		return err
	}

	var in any
	err = ghClient.REST(opts.repo.RepoHost(), "POST", fmt.Sprintf("repos/%s/dispatches", opts.repo.RepoFullName()), &buf, &in)
	if err != nil {
		return err
	}

	wfs, err := getWorkflows(ghClient, opts.repo.RepoHost(), opts.repo.RepoFullName())
	if err != nil {
		return err
	}

	var workflowID int64
	for _, wf := range wfs {
		if wf.Name == opts.workflow {
			workflowID = wf.ID
			break
		}
	}

	runID, err := getRunID(ghClient, opts.repo, "repository_dispatch", workflowID)
	if err != nil {
		return err
	}

	run, err := runShared.GetRun(ghClient, opts.repo, fmt.Sprintf("%d", runID), 0)
	if err != nil {
		return fmt.Errorf("failed to get run: %w", err)
	}

	return render(opts.io, ghClient, opts.repo, run)
}

func getWorkflows(client *cliapi.Client, repoHost string, repoFullName string) ([]shared.Workflow, error) {
	perPage := 100
	page := 1
	workflows := []shared.Workflow{}

	for {
		result := shared.WorkflowsPayload{}
		path := fmt.Sprintf("repos/%s/actions/workflows?per_page=%d&page=%d", repoFullName, perPage, page)
		err := client.REST(repoHost, "GET", path, nil, &result)
		if err != nil {
			return nil, err
		}

		workflows = append(workflows, result.Workflows...)
		if len(result.Workflows) < perPage {
			break
		}

		page++
	}

	return workflows, nil
}
