package dispatch

import (
	"bytes"
	"fmt"
	"io"
	"time"

	cliapi "github.com/cli/cli/v2/api"
	"github.com/cli/cli/v2/pkg/cmd/run/shared"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/iostreams"
)

func render(ios *iostreams.IOStreams, client *cliapi.Client, repo *ghRepo, run *shared.Run) error {
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

		// TODO: should the refresh interval be configurable?
		interval := 2
		fmt.Fprintln(ios.Out, cs.Boldf("Refreshing run status every %d seconds. Press Ctrl+C to quit.", interval))
		fmt.Fprintln(ios.Out)
		fmt.Fprintln(ios.Out, cs.Boldf("https://github.com/%s/actions/runs/%d", repo.RepoFullName(), run.ID))
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

// renderRun is largely an emulation of the upstream 'gh run watch' implementation...
// https://github.com/cli/cli/blob/v2.20.2/pkg/cmd/run/watch/watch.go
func renderRun(out io.Writer, cs *iostreams.ColorScheme, client *cliapi.Client, repo *ghRepo, run *shared.Run, annotationCache map[int64][]shared.Annotation) (*shared.Run, error) {
	run, err := shared.GetRun(client, repo, fmt.Sprintf("%d", run.ID), 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get run: %w", err)
	}

	jobs, err := shared.GetJobs(client, repo, run, 0)
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

		as, annotationErr = shared.GetAnnotations(client, repo, job)
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

	fmt.Fprintln(out, shared.RenderRunHeader(cs, *run, "", "", 0))
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

func getRunID(client *cliapi.Client, repo *ghRepo, event string, workflowID int64) (int64, error) {
	actor, err := cliapi.CurrentLoginName(client, repo.RepoHost())
	if err != nil {
		return 0, err
	}

	for {
		runs, err := shared.GetRunsWithFilter(client, repo, &shared.FilterOptions{
			WorkflowID: workflowID,
			Actor:      actor,
		}, 1, func(run shared.Run) bool {
			// TODO: should this try to match on a branch too?
			// https://github.com/cli/cli/blob/trunk/pkg/cmd/run/shared/shared.go#L281
			return run.Status != shared.Completed && run.WorkflowID == workflowID && run.Event == event
		})
		if err != nil {
			return 0, err
		}

		if len(runs) > 0 {
			return runs[0].ID, nil
		}
	}
}
