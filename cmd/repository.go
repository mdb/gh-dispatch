package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
	"github.com/spf13/cobra"
)

type repositoryDispatchOptions struct {
	Repo          string
	ClientPayload interface{}
	EventType     string
	Workflow      string
	IO            *iostreams.IOStreams
	HTTPTransport http.RoundTripper
	AuthToken     string
}

type repositoryDispatchRequest struct {
	EventType     string      `json:"event_type"`
	ClientPayload interface{} `json:"client_payload"`
}

type workflowRun struct {
	ID         int    `json:"id"`
	WorkflowID int    `json:"workflow_id"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
}

type workflowRunsResponse struct {
	WorkflowRuns []workflowRun `json:"workflow_runs"`
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

		return repositoryDispatchRun(&repositoryDispatchOptions{
			ClientPayload: repoClientPayload,
			EventType:     repositoryEventType,
			Workflow:      repositoryWorkflow,
			Repo:          repo,
			HTTPTransport: http.DefaultTransport,
		})
	},
}

func repositoryDispatchRun(opts *repositoryDispatchOptions) error {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(repositoryDispatchRequest{
		EventType:     opts.EventType,
		ClientPayload: opts.ClientPayload,
	})
	if err != nil {
		return err
	}

	client, err := gh.RESTClient(&api.ClientOptions{
		Transport: opts.HTTPTransport,
		AuthToken: opts.AuthToken,
	})
	if err != nil {
		return err
	}

	var in interface{}
	err = client.Post(fmt.Sprintf("repos/%s/dispatches", opts.Repo), &buf, &in)
	if err != nil {
		return err
	}

	if opts.Workflow == "" {
		return nil
	}

	var wRuns workflowRunsResponse
	err = client.Get(fmt.Sprintf("repos/%s/actions/runs?name=%s&event_type=repository_dispatch", opts.Repo, opts.Workflow), &wRuns)
	if err != nil {
		return err
	}

	id := wRuns.WorkflowRuns[0].ID
	var wRun workflowRun
	err = client.Get(fmt.Sprintf("repos/%s/actions/runs/%d/attempts/1", opts.Repo, id), &wRun)
	if err != nil {
		return err
	}

	fmt.Println(wRun)

	return nil
}

func init() {
	repositoryCmd.Flags().StringVarP(&repositoryEventType, "event-type", "e", "", "The repository dispatch event type.")
	repositoryCmd.MarkFlagRequired("event-type")
	repositoryCmd.Flags().StringVarP(&repositoryClientPayload, "client-payload", "p", "", "The repository dispatch event client payload JSON string.")
	repositoryCmd.MarkFlagRequired("client-payload")
	repositoryCmd.Flags().StringVarP(&repositoryWorkflow, "workflow", "w", "", "The resulting GitHub Actions workflow name.")

	rootCmd.AddCommand(repositoryCmd)
}
