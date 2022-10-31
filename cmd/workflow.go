package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cli/cli/v2/pkg/cmd/run/shared"
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

	runID, err := getWorkflowDispatchRunID(client, opts.Repo, opts.Workflow)
	if err != nil {
		return err
	}

	run, err := getRun(client, opts.Repo, runID)
	if err != nil {
		return fmt.Errorf("failed to get run: %w", err)
	}

	cs := opts.IO.ColorScheme()
	annotationCache := map[int64][]shared.Annotation{}
	out := &bytes.Buffer{}
	opts.IO.StartAlternateScreenBuffer()

	for run.Status != shared.Completed {
		// Write to a temporary buffer to reduce total number of fetches
		run, err = renderRun(out, opts.IO, client, opts.Repo, run, annotationCache)
		if err != nil {
			return err
		}

		if run.Status == shared.Completed {
			break
		}

		// If not completed, refresh the screen buffer and write the temporary buffer to stdout
		opts.IO.RefreshScreen()

		interval := 3
		fmt.Fprintln(opts.IO.Out, cs.Boldf("Refreshing run status every %d seconds. Press Ctrl+C to quit.", interval))
		fmt.Fprintln(opts.IO.Out)
		fmt.Fprintln(opts.IO.Out, cs.Boldf("https://github.com/%s/actions/runs/%d", opts.Repo, runID))
		fmt.Fprintln(opts.IO.Out)

		_, err = io.Copy(opts.IO.Out, out)
		out.Reset()
		if err != nil {
			break
		}

		duration, err := time.ParseDuration(fmt.Sprintf("%ds", interval))
		if err != nil {
			return fmt.Errorf("could not parse interval: %w", err)
		}
		time.Sleep(duration)
	}

	opts.IO.StopAlternateScreenBuffer()

	return nil
}

func getWorkflowDispatchRunID(client api.RESTClient, repo, workflow string) (int64, error) {
	for {
		var wRuns workflowRunsResponse
		err := client.Get(fmt.Sprintf("repos/%s/actions/runs?name=%s&event=workflow_dispatch", repo, workflow), &wRuns)
		if err != nil {
			return 0, err
		}

		if wRuns.WorkflowRuns[0].Status != shared.Completed {
			return wRuns.WorkflowRuns[0].ID, nil
		}
	}
}

func init() {
	workflowCmd.Flags().StringVarP(&workflowInputs, "inputs", "i", "", "The workflow dispatch inputs JSON string.")
	workflowCmd.MarkFlagRequired("inputs")
	workflowCmd.Flags().StringVarP(&workflowRef, "ref", "f", "main", "The git reference for the workflow. Can be a branch or tag name.")
	workflowCmd.Flags().StringVarP(&workflowName, "workflow", "w", "", "The resulting GitHub Actions workflow name.")

	rootCmd.AddCommand(workflowCmd)
}
