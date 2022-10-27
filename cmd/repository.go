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
	ID         int64  `json:"id"`
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

		ios := iostreams.System()

		return repositoryDispatchRun(&repositoryDispatchOptions{
			ClientPayload: repoClientPayload,
			EventType:     repositoryEventType,
			Workflow:      repositoryWorkflow,
			Repo:          repo,
			HTTPTransport: http.DefaultTransport,
			IO:            ios,
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

	// TODO: is there a more accurate way to fetch the workflow run without sleeping?
	// See https://docs.github.com/en/rest/actions/workflow-runs#list-workflow-runs-for-a-repository
	sleep, err := time.ParseDuration("5s")
	if err != nil {
		return fmt.Errorf("could not parse interval: %w", err)
	}
	time.Sleep(sleep)

	var wRuns workflowRunsResponse
	path := fmt.Sprintf("repos/%s/actions/runs?name=%s&event=repository_dispatch", opts.Repo, opts.Workflow)
	err = client.Get(path, &wRuns)
	if err != nil {
		return err
	}

	runID := wRuns.WorkflowRuns[0].ID
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
		run, err = renderRun(out, *opts, client, opts.Repo, run, annotationCache)
		if err != nil {
			break
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

func renderRun(out io.Writer, opts repositoryDispatchOptions, client api.RESTClient, repo string, run *shared.Run, annotationCache map[int64][]shared.Annotation) (*shared.Run, error) {
	cs := opts.IO.ColorScheme()

	var err error

	run, err = getRun(client, repo, run.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get run: %w", err)
	}

	jobs, err := getJobs(client, repo, run.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs: %w", err)
	}

	var annotations []shared.Annotation

	var annotationErr error
	var as []shared.Annotation
	for _, job := range jobs {
		if as, ok := annotationCache[job.ID]; ok {
			annotations = as
			continue
		}

		as, annotationErr = getAnnotations(client, repo, job)
		if annotationErr != nil {
			break
		}
		annotations = append(annotations, as...)

		if job.Status != shared.InProgress {
			annotationCache[job.ID] = annotations
		}
	}

	if annotationErr != nil {
		return nil, fmt.Errorf("failed to get annotations: %w", annotationErr)
	}

	fmt.Fprintln(out)

	if len(jobs) == 0 {
		return run, nil
	}

	fmt.Fprintln(out, cs.Bold("JOBS"))

	fmt.Fprintln(out, shared.RenderJobs(cs, jobs, true))

	if len(annotations) > 0 {
		fmt.Fprintln(out)
		fmt.Fprintln(out, cs.Bold("ANNOTATIONS"))
		fmt.Fprintln(out, shared.RenderAnnotations(cs, annotations))
	}

	return run, nil
}

func getRun(client api.RESTClient, repo string, runID int64) (*shared.Run, error) {
	var result shared.Run
	err := client.Get(fmt.Sprintf("repos/%s/actions/runs/%d", repo, runID), &result)
	if err != nil {
		return nil, err
	}

	// Set name to workflow name
	/*
		workflow, err := workflowShared.GetWorkflow(client, repo, result.WorkflowID)
		if err != nil {
			return nil, err
		} else {
			result.WorkflowName = workflow.Name
		}
	*/

	return &result, nil
}

func getJobs(client api.RESTClient, repo string, runID int64) ([]shared.Job, error) {
	var result shared.JobsPayload
	err := client.Get(fmt.Sprintf("repos/%s/actions/runs/%d/attempts/1/jobs", repo, runID), &result)
	if err != nil {
		return nil, err
	}

	return result.Jobs, nil
}

func getAnnotations(client api.RESTClient, repo string, job shared.Job) ([]shared.Annotation, error) {
	var result []*shared.Annotation
	err := client.Get(fmt.Sprintf("repos/%s/check-runs/%d/annotations", repo, job.ID), &result)
	if err != nil {
		return nil, err
	}

	out := []shared.Annotation{}
	for _, annotation := range result {
		annotation.JobName = job.Name
		out = append(out, *annotation)
	}

	return out, nil
}

func init() {
	repositoryCmd.Flags().StringVarP(&repositoryEventType, "event-type", "e", "", "The repository dispatch event type.")
	repositoryCmd.MarkFlagRequired("event-type")
	repositoryCmd.Flags().StringVarP(&repositoryClientPayload, "client-payload", "p", "", "The repository dispatch event client payload JSON string.")
	repositoryCmd.MarkFlagRequired("client-payload")
	repositoryCmd.Flags().StringVarP(&repositoryWorkflow, "workflow", "w", "", "The resulting GitHub Actions workflow name.")

	rootCmd.AddCommand(repositoryCmd)
}
