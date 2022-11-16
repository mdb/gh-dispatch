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

type workflowDispatchRequest struct {
	Inputs interface{} `json:"inputs"`
	Ref    string      `json:"ref"`
}

type workflowDispatchOptions struct {
	inputs   interface{}
	ref      string
	workflow string
	dispatchOptions
}

func NewCmdWorkflow() *cobra.Command {
	var (
		workflowInputs string
		workflowName   string
		workflowRef    string
	)

	cmd := &cobra.Command{
		Use: heredoc.Doc(`
		workflow \
			--repo [owner/repo] \
			--inputs [json-string] \
			--workflow [workflow-file-name.yaml]
	`),
		Short: "Send a workflow dispatch event and watch the resulting GitHub Actions run",
		Long: heredoc.Doc(`
		This command sends a workflow dispatch event and attempts to find and watch the
		resulting GitHub Actions run whose file name or ID is specified as '--workflow'.

		Note that the command assumes the specified workflow supports a workflow_dispatch
		'on' trigger. Also note that the command is vulnerable to race conditions and may
		watch an unrelated GitHub Actions workflow run in the event that multiple runs of
		the specified workflow are running concurrently.
	`),
		Example: heredoc.Doc(`
		gh dispatch workflow \
			--repo mdb/gh-dispatch \
			--inputs '{"name": "Mike"}' \
			--workflow workflow_dispatch.yaml

		# Specify a workflow ref other than 'main'
		gh dispatch workflow \
			--repo mdb/gh-dispatch \
			--inputs '{"name": "Mike"}' \
			--workflow workflow_dispatch.yaml \
			--ref my-feature-branch
	`),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, _ := cmd.Flags().GetString("repo")

			b := []byte(workflowInputs)
			var wInputs interface{}
			json.Unmarshal(b, &wInputs)

			ios := iostreams.System()

			dOptions := dispatchOptions{
				repo:          repo,
				httpTransport: http.DefaultTransport,
				io:            ios,
			}

			return workflowDispatchRun(&workflowDispatchOptions{
				inputs:          wInputs,
				ref:             workflowRef,
				workflow:        workflowName,
				dispatchOptions: dOptions,
			})
		},
	}

	// TODO: how does the 'gh run' command represent inputs?
	// Is it worth better emulating its interface?
	cmd.Flags().StringVarP(&workflowInputs, "inputs", "i", "", "The workflow dispatch inputs JSON string.")
	cmd.MarkFlagRequired("inputs")
	// TODO: how does the 'gh run' command represent workflow?
	// Is it worth better emulating its interface?
	cmd.Flags().StringVarP(&workflowName, "workflow", "w", "", "The resulting GitHub Actions workflow name.")
	cmd.MarkFlagRequired("workflow")
	// TODO: how does the 'gh run' command represent ref?
	// Is it worth better emulating its interface?
	cmd.Flags().StringVarP(&workflowRef, "ref", "f", "main", "The git reference for the workflow. Can be a branch or tag name.")

	return cmd
}

func workflowDispatchRun(opts *workflowDispatchOptions) error {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(workflowDispatchRequest{
		Inputs: opts.inputs,
		Ref:    opts.ref,
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
	err = client.Post(fmt.Sprintf("repos/%s/actions/workflows/%s/dispatches", opts.repo, opts.workflow), &buf, &in)
	if err != nil {
		return err
	}

	var wf shared.Workflow
	err = client.Get(fmt.Sprintf("repos/%s/actions/workflows/%s", opts.repo, opts.workflow), &wf)
	if err != nil {
		return err
	}

	ghClient := cliapi.NewClientFromHTTP(&http.Client{
		Transport: opts.httpTransport,
	})

	runID, err := getRunID(ghClient, opts.repo, "workflow_dispatch", wf.ID)
	if err != nil {
		return err
	}

	run, err := getRun(client, opts.repo, runID)
	if err != nil {
		return fmt.Errorf("failed to get run: %w", err)
	}

	return render(opts.io, client, opts.repo, run)
}
