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

type repositoryDispatchRequest struct {
	EventType     string      `json:"event_type"`
	ClientPayload interface{} `json:"client_payload"`
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

		return repositoryDispatchRun(&dispatchOptions{
			ClientPayload: repoClientPayload,
			EventType:     repositoryEventType,
			Workflow:      repositoryWorkflow,
			Repo:          repo,
			HTTPTransport: http.DefaultTransport,
			IO:            ios,
		})
	},
}

func repositoryDispatchRun(opts *dispatchOptions) error {
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

	runID, err := getRunID(client, opts.Repo, opts.Workflow)
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
		run, err = renderRun(out, *opts, client, opts.Repo, run, annotationCache)
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

func init() {
	repositoryCmd.Flags().StringVarP(&repositoryEventType, "event-type", "e", "", "The repository dispatch event type.")
	repositoryCmd.MarkFlagRequired("event-type")
	repositoryCmd.Flags().StringVarP(&repositoryClientPayload, "client-payload", "p", "", "The repository dispatch event client payload JSON string.")
	repositoryCmd.MarkFlagRequired("client-payload")
	repositoryCmd.Flags().StringVarP(&repositoryWorkflow, "workflow", "w", "", "The resulting GitHub Actions workflow name.")

	rootCmd.AddCommand(repositoryCmd)
}
