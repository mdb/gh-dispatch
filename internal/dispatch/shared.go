package dispatch

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	cliapi "github.com/cli/cli/v2/api"
	"github.com/cli/cli/v2/pkg/cmd/run/shared"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/cli/go-gh/pkg/api"
)

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
	repo          string
	httpTransport http.RoundTripper
	io            *iostreams.IOStreams
	authToken     string
}

type ghRepo struct {
	Name  string
	Owner string
}

func (r ghRepo) RepoName() string {
	return r.Name
}

func (r ghRepo) RepoOwner() string {
	return r.Owner
}

func (r ghRepo) RepoHost() string {
	return "github.com"
}

func getRunID(client *cliapi.Client, repo, event string, workflowID int64) (int64, error) {
	repoParts := strings.Split(repo, "/")
	r := &ghRepo{
		Owner: repoParts[0],
		Name:  repoParts[1],
	}

	for {
		runs, err := shared.GetRuns(client, r, &shared.FilterOptions{
			// TODO: should we detect/pass an Actor as well? Perhaps a Branch too?
			// Alternatively, should FilterOptions have an Event field?
			// https://github.com/cli/cli/blob/trunk/pkg/cmd/run/shared/shared.go#L281
			WorkflowID: workflowID,
		}, 100)
		if err != nil {
			return 0, err
		}

		// TODO: match on workflow name, or somehow more accurately ensure we are fetching
		// _the_ workflow triggered by the `gh dispatch` command.
		// TODO: also match on event
		for _, run := range runs.WorkflowRuns {
			// TODO: should this also try to match on run.triggering_actor.login?
			if run.Status != shared.Completed && run.WorkflowID == workflowID {
				return run.ID, nil
			}
		}
	}
}

func render(ios *iostreams.IOStreams, client api.RESTClient, repo string, run *shared.Run) error {
	cs := ios.ColorScheme()
	annotationCache := map[int64][]shared.Annotation{}
	out := &bytes.Buffer{}
	ios.StartAlternateScreenBuffer()

	for run.Status != shared.Completed {
		// Write to a temporary buffer to reduce total number of fetches
		var err error
		run, err = renderRun(out, cs, client, repo, run, annotationCache)
		if err != nil {
			return err
		}

		// If not completed, refresh the screen buffer and write the temporary buffer to stdout
		ios.RefreshScreen()

		interval := 2
		fmt.Fprintln(ios.Out, cs.Boldf("Refreshing run status every %d seconds. Press Ctrl+C to quit.", interval))
		fmt.Fprintln(ios.Out)
		fmt.Fprintln(ios.Out, cs.Boldf("https://github.com/%s/actions/runs/%d", repo, run.ID))
		fmt.Fprintln(ios.Out)

		_, err = io.Copy(ios.Out, out)
		out.Reset()
		if err != nil {
			return err
		}

		duration, err := time.ParseDuration(fmt.Sprintf("%ds", interval))
		if err != nil {
			return fmt.Errorf("could not parse interval: %w", err)
		}
		time.Sleep(duration)
	}

	ios.StopAlternateScreenBuffer()

	symbol, symbolColor := shared.Symbol(cs, run.Status, run.Conclusion)
	id := cs.Cyanf("%d", run.ID)

	if ios.IsStdoutTTY() {
		fmt.Fprintln(ios.Out)
		fmt.Fprintf(ios.Out, "%s %s (%s) completed with '%s'\n", symbolColor(symbol), cs.Bold(run.Name), id, run.Conclusion)
	}

	if run.Conclusion != shared.Success {
		return cmdutil.SilentError
	}

	return nil
}

func renderRun(out io.Writer, cs *iostreams.ColorScheme, client api.RESTClient, repo string, run *shared.Run, annotationCache map[int64][]shared.Annotation) (*shared.Run, error) {
	run, err := getRun(client, repo, run.ID)
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

	//fmt.Fprintln(out, shared.RenderRunHeader(cs, *run, "", ""))
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

func getRun2(client *cliapi.Client, repo string, runID int64) (*shared.Run, error) {
	repoParts := strings.Split(repo, "/")
	r := &ghRepo{
		Owner: repoParts[0],
		Name:  repoParts[1],
	}

	return shared.GetRun(client, r, fmt.Sprintf("%d", runID))
}

func getRun(client api.RESTClient, repo string, runID int64) (*shared.Run, error) {
	var result shared.Run
	err := client.Get(fmt.Sprintf("repos/%s/actions/runs/%d", repo, runID), &result)
	if err != nil {
		return nil, err
	}

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

	annotations := []shared.Annotation{}
	for _, annotation := range result {
		annotation.JobName = job.Name
		annotations = append(annotations, *annotation)
	}

	return annotations, nil
}
