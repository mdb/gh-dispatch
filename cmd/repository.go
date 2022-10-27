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
	IO            *iostreams.IOStreams
	HTTPTransport http.RoundTripper
	AuthToken     string
}

type repositoryDispatchRequest struct {
	EventType     string      `json:"event_type"`
	ClientPayload interface{} `json:"client_payload"`
}

var (
	repositoryEventType     string
	repositoryClientPayload string
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

	return nil
}

func init() {
	repositoryCmd.Flags().StringVarP(&repositoryEventType, "event-type", "e", "", "The repository dispatch event type.")
	repositoryCmd.MarkFlagRequired("event-type")
	repositoryCmd.Flags().StringVarP(&repositoryClientPayload, "client-payload", "p", "", "The repository dispatch event client payload JSON string.")
	repositoryCmd.MarkFlagRequired("client-payload")

	rootCmd.AddCommand(repositoryCmd)
}
