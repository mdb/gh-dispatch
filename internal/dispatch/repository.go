package dispatch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/MakeNowJust/heredoc"
	cliapi "github.com/cli/cli/v2/api"
	"github.com/cli/cli/v2/pkg/cmd/workflow/shared"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
	"github.com/spf13/cobra"
)

type repositoryDispatchRequest struct {
	EventType     string      `json:"event_type"`
	ClientPayload interface{} `json:"client_payload"`
}

type repositoryDispatchOptions struct {
	clientPayload interface{}
	eventType     string
	workflow      string
	dispatchOptions
}

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
			repo, _ := cmd.Flags().GetString("repo")

			b := []byte(repositoryClientPayload)
			var repoClientPayload interface{}
			json.Unmarshal(b, &repoClientPayload)

			ios := iostreams.System()

			dOptions := dispatchOptions{
				repo:          repo,
				httpTransport: http.DefaultTransport,
				io:            ios,
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
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(repositoryDispatchRequest{
		EventType:     opts.eventType,
		ClientPayload: opts.clientPayload,
	})
	if err != nil {
		return err
	}

	client, err := gh.RESTClient(&api.ClientOptions{
		Transport: opts.httpTransport,
		AuthToken: opts.authToken,
	})
	if err != nil {
		return err
	}

	var in interface{}
	err = client.Post(fmt.Sprintf("repos/%s/dispatches", opts.repo), &buf, &in)
	if err != nil {
		return err
	}

	var wfs shared.WorkflowsPayload
	err = client.Get(fmt.Sprintf("repos/%s/actions/workflows", opts.repo), &wfs)
	if err != nil {
		return err
	}

	var workflowID int64
	for _, wf := range wfs.Workflows {
		if wf.Name == opts.workflow {
			workflowID = wf.ID
			break
		}
	}

	httpClient := &http.Client{
		Transport: opts.httpTransport,
	}
	ghClient := cliapi.NewClientFromHTTP(httpClient)

	runID, err := getRunID2(ghClient, opts.repo, "repository_dispatch", workflowID)
	if err != nil {
		return err
	}

	run, err := getRun(client, opts.repo, runID)
	if err != nil {
		return fmt.Errorf("failed to get run: %w", err)
	}

	return render(opts.io, client, opts.repo, run)
}