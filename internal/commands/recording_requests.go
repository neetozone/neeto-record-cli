package commands

import (
	"github.com/spf13/cobra"
)

var recordingRequestsCmd = &cobra.Command{
	Use:   "recording-requests",
	Short: "Manage recording requests",
}

var recordingRequestsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a recording request",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		title, _ := cmd.Flags().GetString("title")
		createdByEmail, _ := cmd.Flags().GetString("created-by-email")

		recording := map[string]interface{}{
			"title":            title,
			"created_by_email": createdByEmail,
		}

		if v, _ := cmd.Flags().GetString("request-instructions"); cmd.Flags().Changed("request-instructions") {
			recording["request_instructions"] = v
		}
		if v, _ := cmd.Flags().GetString("request-notes"); cmd.Flags().Changed("request-notes") {
			recording["request_notes"] = v
		}

		data, err := c.Post("/recording_requests", map[string]interface{}{"recording": recording})
		if err != nil {
			return err
		}

		printActionResult(data, nil)
		return nil
	},
}

func init() {
	recordingRequestsCreateCmd.Flags().String("title", "", "Title of the requested recording")
	recordingRequestsCreateCmd.Flags().String("created-by-email", "", "Email of the recording requester")
	recordingRequestsCreateCmd.Flags().String(
		"request-instructions",
		"",
		"Instructions shown to the person who will upload the recording",
	)
	recordingRequestsCreateCmd.Flags().String("request-notes", "", "Private notes for this request")
	_ = recordingRequestsCreateCmd.MarkFlagRequired("title")
	_ = recordingRequestsCreateCmd.MarkFlagRequired("created-by-email")

	recordingRequestsCmd.AddCommand(recordingRequestsCreateCmd)
	rootCmd.AddCommand(recordingRequestsCmd)
}
