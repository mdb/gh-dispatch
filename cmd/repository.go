package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

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

var (
	repositoryEventType     string
	repositoryClientPayload string
	repositoryWorkflow      string
)

// repositoryCmd represents the repository subcommand
var repositoryCmd = &cobra.Command{
	Use:     "repository",
	Short:   `The 'repository' subcommand triggers repository dispatch events`,
	Long:    `The 'repository' subcommand triggers repository dispatch events`,
	Example: `TODO`,
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

	runID, err := getRunID(client, opts.repo, "repository_dispatch", workflowID)
	if err != nil {
		return err
	}

	run, err := getRun(client, opts.repo, runID)
	if err != nil {
		return fmt.Errorf("failed to get run: %w", err)
	}

	return render(opts.io, client, opts.repo, run)
}

func init() {
	repositoryCmd.Flags().StringVarP(&repositoryEventType, "event-type", "e", "", "The repository dispatch event type.")
	repositoryCmd.MarkFlagRequired("event-type")
	repositoryCmd.Flags().StringVarP(&repositoryClientPayload, "client-payload", "p", "", "The repository dispatch event client payload JSON string.")
	repositoryCmd.MarkFlagRequired("client-payload")
	repositoryCmd.Flags().StringVarP(&repositoryWorkflow, "workflow", "w", "", "The resulting GitHub Actions workflow name.")
	repositoryCmd.MarkFlagRequired("workflow")

	rootCmd.AddCommand(repositoryCmd)
}
