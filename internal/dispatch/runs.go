package dispatch

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/cli/go-gh/v2/pkg/api"
)

// The types and functions in this file are a subset of those found in the
// upstream github.com/cli/cli run/workflow shared packages, adapted to
// use the go-gh REST client.

type status string
type conclusion string
type annotationLevel string

const (
	statusCompleted  status = "completed"
	statusInProgress status = "in_progress"

	conclusionSuccess conclusion = "success"
	conclusionSkipped conclusion = "skipped"
	conclusionNeutral conclusion = "neutral"

	annotationFailure annotationLevel = "failure"
	annotationWarning annotationLevel = "warning"
)

var errMissingAnnotationsPermissions = errors.New("missing annotations permissions error")

type workflow struct {
	ID   int64
	Name string
}

type workflowRun struct {
	ID           int64
	Name         string
	Status       status
	Conclusion   conclusion
	Event        string
	CreatedAt    time.Time `json:"created_at"`
	WorkflowID   int64     `json:"workflow_id"`
	HeadBranch   string    `json:"head_branch"`
	JobsURL      string    `json:"jobs_url"`
	WorkflowName string    `json:"-"`
}

type job struct {
	ID          int64
	Status      status
	Conclusion  conclusion
	Name        string
	Steps       []step
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at"`
}

type step struct {
	Name       string
	Status     status
	Conclusion conclusion
}

type annotation struct {
	JobName   string
	Message   string
	Path      string
	Level     annotationLevel `json:"annotation_level"`
	StartLine int             `json:"start_line"`
}

func currentLoginName(client *api.RESTClient) (string, error) {
	var user struct {
		Login string
	}
	err := client.Get("user", &user)
	return user.Login, err
}

func getWorkflow(client *api.RESTClient, repo *ghRepo, workflowIDOrFile string) (*workflow, error) {
	var wf workflow
	path := fmt.Sprintf("repos/%s/actions/workflows/%s", repo.RepoFullName(), workflowIDOrFile)
	if err := client.Get(path, &wf); err != nil {
		return nil, err
	}

	return &wf, nil
}

func getWorkflows(client *api.RESTClient, repo *ghRepo) ([]workflow, error) {
	perPage := 100
	page := 1
	workflows := []workflow{}

	for {
		var result struct {
			Workflows []workflow
		}
		path := fmt.Sprintf("repos/%s/actions/workflows?per_page=%d&page=%d", repo.RepoFullName(), perPage, page)
		if err := client.Get(path, &result); err != nil {
			return nil, err
		}

		workflows = append(workflows, result.Workflows...)
		if len(result.Workflows) < perPage {
			break
		}

		page++
	}

	return workflows, nil
}

// getRuns fetches the 50 most recent runs of the workflow triggered by
// the actor and returns those for which filter returns true.
func getRuns(client *api.RESTClient, repo *ghRepo, workflowID int64, actor string, filter func(workflowRun) bool) ([]workflowRun, error) {
	path := fmt.Sprintf("repos/%s/actions/runs", repo.RepoFullName())
	if workflowID > 0 {
		path = fmt.Sprintf("repos/%s/actions/workflows/%d/runs", repo.RepoFullName(), workflowID)
	}

	query := url.Values{}
	query.Set("per_page", "50")
	query.Set("exclude_pull_requests", "true")
	if actor != "" {
		query.Set("actor", actor)
	}

	var result struct {
		WorkflowRuns []workflowRun `json:"workflow_runs"`
	}
	if err := client.Get(fmt.Sprintf("%s?%s", path, query.Encode()), &result); err != nil {
		return nil, err
	}

	var filtered []workflowRun
	for _, r := range result.WorkflowRuns {
		if filter(r) {
			filtered = append(filtered, r)
		}
	}

	return filtered, nil
}

func getRun(client *api.RESTClient, repo *ghRepo, runID int64) (*workflowRun, error) {
	var r workflowRun
	path := fmt.Sprintf("repos/%s/actions/runs/%d?exclude_pull_requests=true", repo.RepoFullName(), runID)
	if err := client.Get(path, &r); err != nil {
		return nil, err
	}

	wf, err := getWorkflow(client, repo, fmt.Sprintf("%d", r.WorkflowID))
	if err != nil {
		return nil, err
	}

	r.WorkflowName = wf.Name

	return &r, nil
}

func getJobs(client *api.RESTClient, r *workflowRun) ([]job, error) {
	query := url.Values{}
	query.Set("per_page", "100")
	path := fmt.Sprintf("%s?%s", r.JobsURL, query.Encode())

	jobs := []job{}
	for path != "" {
		resp, err := client.Request(http.MethodGet, path, nil)
		if err != nil {
			return nil, err
		}

		var payload struct {
			Jobs []job
		}
		err = json.NewDecoder(resp.Body).Decode(&payload)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		jobs = append(jobs, payload.Jobs...)
		path = findNextPage(resp)
	}

	return jobs, nil
}

var linkRE = regexp.MustCompile(`<([^>]+)>;\s*rel="([^"]+)"`)

func findNextPage(resp *http.Response) string {
	for _, m := range linkRE.FindAllStringSubmatch(resp.Header.Get("Link"), -1) {
		if len(m) > 2 && m[2] == "next" {
			return m[1]
		}
	}
	return ""
}

// getAnnotations returns an empty slice if the job has no annotations, and
// errMissingAnnotationsPermissions if the API responds with a 403.
func getAnnotations(client *api.RESTClient, repo *ghRepo, j job) ([]annotation, error) {
	var result []*annotation
	path := fmt.Sprintf("repos/%s/check-runs/%d/annotations", repo.RepoFullName(), j.ID)

	if err := client.Get(path, &result); err != nil {
		var httpErr *api.HTTPError
		if !errors.As(err, &httpErr) {
			return nil, err
		}

		switch httpErr.StatusCode {
		case http.StatusNotFound:
			return []annotation{}, nil
		case http.StatusForbidden:
			return nil, errMissingAnnotationsPermissions
		default:
			return nil, err
		}
	}

	out := []annotation{}
	for _, a := range result {
		a.JobName = j.Name
		out = append(out, *a)
	}

	return out, nil
}
