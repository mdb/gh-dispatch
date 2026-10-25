package dispatch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/cli/go-gh/v2/pkg/api"
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

			client, err := api.NewRESTClient(api.ClientOptions{Host: repo.RepoHost()})
			if err != nil {
				return err
			}
			dOptions := dispatchOptions{
				repo:   repo,
				client: client,
				io:     iostreams.System(),
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
	client := opts.client

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(repositoryDispatchRequest{
		EventType:     opts.eventType,
		ClientPayload: opts.clientPayload,
	})
	if err != nil {
		return err
	}

	var in any
	dispatchedAt := time.Now()
	err = client.Post(fmt.Sprintf("repos/%s/dispatches", opts.repo.RepoFullName()), &buf, &in)
	if err != nil {
		return err
	}

	wfs, err := getWorkflows(client, opts.repo)
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

	runID, err := getRunID(client, opts.repo, "repository_dispatch", workflowID, dispatchedAt)
	if err != nil {
		return err
	}

	run, err := getRun(client, opts.repo, runID)
	if err != nil {
		return fmt.Errorf("failed to get run: %w", err)
	}

	return render(opts.io, client, opts.repo, run)
}
