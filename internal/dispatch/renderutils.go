package dispatch

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/cli/go-gh/v2/pkg/api"
)

func render(ios *iostreams.IOStreams, client *api.RESTClient, repo *ghRepo, run *workflowRun) error {
	cs := ios.ColorScheme()
	annotationCache := map[int64][]annotation{}
	out := &bytes.Buffer{}
	ios.StartAlternateScreenBuffer()

	for {
		// Write to a temporary buffer to reduce total number of fetches
		var err error
		run, err = renderRun(out, cs, client, repo, run, annotationCache)
		if err != nil {
			return err
		}

		// Refresh the screen buffer and write the temporary buffer to stdout
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

		if run.Status == statusCompleted {
			break
		}

		duration, err := time.ParseDuration(fmt.Sprintf("%ds", interval))
		if err != nil {
			return fmt.Errorf("could not parse interval: %w", err)
		}
		time.Sleep(duration)
	}

	ios.StopAlternateScreenBuffer()

	symbol, symbolColor := renderSymbol(cs, run.Status, run.Conclusion)
	id := cs.Cyanf("%d", run.ID)

	if ios.IsStdoutTTY() {
		fmt.Fprintln(ios.Out)
		fmt.Fprintf(ios.Out, "%s %s (%s) completed with '%s'\n", symbolColor(symbol), cs.Bold(run.Name), id, run.Conclusion)
	}

	if run.Conclusion != conclusionSuccess {
		return cmdutil.SilentError
	}

	return nil
}

// renderRun is largely an emulation of the upstream 'gh run watch' implementation...
// https://github.com/cli/cli/blob/v2.20.2/pkg/cmd/run/watch/watch.go
func renderRun(out io.Writer, cs *iostreams.ColorScheme, client *api.RESTClient, repo *ghRepo, run *workflowRun, annotationCache map[int64][]annotation) (*workflowRun, error) {
	run, err := getRun(client, repo, run.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get run: %w", err)
	}

	jobs, err := getJobs(client, run)
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs: %w", err)
	}

	var annotations []annotation
	var annotationErr error
	var as []annotation
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

		if job.Status != statusInProgress {
			annotationCache[job.ID] = annotations
		}
	}

	if annotationErr != nil {
		return nil, fmt.Errorf("failed to get annotations: %w", annotationErr)
	}

	fmt.Fprintln(out, renderRunHeader(cs, run))
	fmt.Fprintln(out)

	if len(jobs) == 0 {
		return run, nil
	}

	fmt.Fprintln(out, cs.Bold("JOBS"))
	fmt.Fprintln(out, renderJobs(cs, jobs))

	if len(annotations) > 0 {
		fmt.Fprintln(out)
		fmt.Fprintln(out, cs.Bold("ANNOTATIONS"))
		fmt.Fprintln(out, renderAnnotations(cs, annotations))
	}

	return run, nil
}

func getRunID(client *api.RESTClient, repo *ghRepo, event string, workflowID int64, dispatchedAt time.Time) (int64, error) {
	actor, err := currentLoginName(client)
	if err != nil {
		return 0, err
	}

	for {
		runs, err := getRuns(client, repo, workflowID, actor, func(run workflowRun) bool {
			// TODO: should this try to match on a branch too?
			// https://github.com/cli/cli/blob/trunk/pkg/cmd/run/shared/shared.go#L281
			return run.WorkflowID == workflowID && run.Event == event && !run.CreatedAt.Before(dispatchedAt)
		})
		if err != nil {
			return 0, err
		}

		if len(runs) > 0 {
			return runs[0].ID, nil
		}
	}
}

// The following render functions are adapted from the upstream
// github.com/cli/cli pkg/cmd/run/shared package.

func renderRunHeader(cs *iostreams.ColorScheme, run *workflowRun) string {
	symbol, symbolColor := renderSymbol(cs, run.Status, run.Conclusion)
	title := fmt.Sprintf("%s %s", cs.Bold(run.HeadBranch), run.WorkflowName)
	id := cs.Cyanf("%d", run.ID)

	return fmt.Sprintf("%s %s · %s\nTriggered via %s ", symbolColor(symbol), title, id, run.Event)
}

func renderJobs(cs *iostreams.ColorScheme, jobs []job) string {
	lines := []string{}
	for _, j := range jobs {
		elapsed := j.CompletedAt.Sub(j.StartedAt)
		elapsedStr := fmt.Sprintf(" in %s", elapsed)
		if elapsed < 0 {
			elapsedStr = ""
		}
		symbol, symbolColor := renderSymbol(cs, j.Status, j.Conclusion)
		id := cs.Cyanf("%d", j.ID)
		lines = append(lines, fmt.Sprintf("%s %s%s (ID %s)", symbolColor(symbol), cs.Bold(j.Name), elapsedStr, id))
		for _, s := range j.Steps {
			stepSymbol, stepSymbolColor := renderSymbol(cs, s.Status, s.Conclusion)
			lines = append(lines, fmt.Sprintf("  %s %s", stepSymbolColor(stepSymbol), s.Name))
		}
	}

	return strings.Join(lines, "\n")
}

func renderAnnotations(cs *iostreams.ColorScheme, annotations []annotation) string {
	lines := []string{}
	for _, a := range annotations {
		lines = append(lines, fmt.Sprintf("%s %s", renderAnnotationSymbol(cs, a), a.Message))
		// Following newline is essential for spacing between annotations
		lines = append(lines, cs.Mutedf("%s: %s#%d\n", a.JobName, a.Path, a.StartLine))
	}

	return strings.Join(lines, "\n")
}

func renderSymbol(cs *iostreams.ColorScheme, s status, c conclusion) (string, func(string) string) {
	noColor := func(s string) string { return s }
	if s == statusCompleted {
		switch c {
		case conclusionSuccess:
			return cs.SuccessIconWithColor(noColor), cs.Green
		case conclusionSkipped, conclusionNeutral:
			return "-", cs.Muted
		default:
			return cs.FailureIconWithColor(noColor), cs.Red
		}
	}

	return "*", cs.Yellow
}

func renderAnnotationSymbol(cs *iostreams.ColorScheme, a annotation) string {
	switch a.Level {
	case annotationFailure:
		return cs.FailureIcon()
	case annotationWarning:
		return cs.WarningIcon()
	default:
		return "-"
	}
}
