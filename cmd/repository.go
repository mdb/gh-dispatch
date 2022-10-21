package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/cli/go-gh"
	"github.com/spf13/cobra"
)

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

		var buf bytes.Buffer
		err := json.NewEncoder(&buf).Encode(repositoryDispatchRequest{
			EventType: repositoryEventType,
			// TODO: how to handle JSON of unknown structure?
			ClientPayload: repoClientPayload,
		})
		if err != nil {
			return err
		}

		client, err := gh.RESTClient(nil)
		if err != nil {
			return err
		}

		// TODO: pass a real interface
		err = client.Post(fmt.Sprintf("repos/%s/dispatches", repo), &buf, &repositoryDispatchRequest{})
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	repositoryCmd.Flags().StringVarP(&repositoryEventType, "event-type", "e", "", "The repository dispatch event type.")
	repositoryCmd.MarkFlagRequired("event-type")
	repositoryCmd.Flags().StringVarP(&repositoryClientPayload, "client-payload", "p", "", "The repository dispatch event client payload JSON string.")
	repositoryCmd.MarkFlagRequired("client-payload")

	rootCmd.AddCommand(repositoryCmd)
}
