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

type workflowDispatchRequest struct {
	Inputs interface{} `json:"inputs"`
	Ref    string      `json:"ref"`
}

type workflowDispatchOptions struct {
	Repo          string
	Inputs        interface{}
	Ref           string
	Workflow      string
	IO            *iostreams.IOStreams
	HTTPTransport http.RoundTripper
	AuthToken     string
}

var (
	workflowInputs string
	workflowName   string
	workflowRef    string
)

// workflowCmd represents the workflow subcommand
var workflowCmd = &cobra.Command{
	Use:     "workflow",
	Short:   `The 'workflow' subcommand triggers workflow dispatch events`,
	Long:    `The 'workflow' subcommand triggers workflow dispatch events`,
	Example: `TODO`,
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, _ := cmd.Flags().GetString("repo")

		b := []byte(workflowInputs)
		var wInputs interface{}
		json.Unmarshal(b, &wInputs)

		ios := iostreams.System()

		return workflowDispatchRun(&workflowDispatchOptions{
			Inputs:        wInputs,
			Ref:           workflowRef,
			Workflow:      workflowName,
			Repo:          repo,
			HTTPTransport: http.DefaultTransport,
			IO:            ios,
		})
	},
}

func workflowDispatchRun(opts *workflowDispatchOptions) error {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(workflowDispatchRequest{
		Inputs: opts.Inputs,
		Ref:    opts.Ref,
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
	err = client.Post(fmt.Sprintf("repos/%s/actions/workflows/%s/dispatches", opts.Repo, opts.Workflow), &buf, &in)
	if err != nil {
		return err
	}

	runID, err := getRunID(client, opts.Repo, "workflow_dispatch")
	if err != nil {
		return err
	}

	run, err := getRun(client, opts.Repo, runID)
	if err != nil {
		return fmt.Errorf("failed to get run: %w", err)
	}

	return render(opts.IO, client, opts.Repo, run)
}

func init() {
	workflowCmd.Flags().StringVarP(&workflowInputs, "inputs", "i", "", "The workflow dispatch inputs JSON string.")
	workflowCmd.MarkFlagRequired("inputs")
	workflowCmd.Flags().StringVarP(&workflowRef, "ref", "f", "main", "The git reference for the workflow. Can be a branch or tag name.")
	workflowCmd.Flags().StringVarP(&workflowName, "workflow", "w", "", "The resulting GitHub Actions workflow name.")

	rootCmd.AddCommand(workflowCmd)
}
