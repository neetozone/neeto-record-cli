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

		recordingRequest := map[string]interface{}{"title": title}

		if v, _ := cmd.Flags().GetString("created-by-email"); cmd.Flags().Changed("created-by-email") {
			recordingRequest["created_by_email"] = v
		}

		if v, _ := cmd.Flags().GetString("request-instructions"); cmd.Flags().Changed("request-instructions") {
			recordingRequest["request_instructions"] = v
		}
		if v, _ := cmd.Flags().GetString("request-notes"); cmd.Flags().Changed("request-notes") {
			recordingRequest["request_notes"] = v
		}

		data, err := c.Post("/recording_requests", map[string]interface{}{"recording": recordingRequest})
		if err != nil {
			return err
		}

		printActionResult(data, nil)
		return nil
	},
}

func init() {
	recordingRequestsCreateCmd.Flags().String("title", "", "Title of the requested recording")
	recordingRequestsCreateCmd.Flags().String("created-by-email", "", "Email of the member who owns the request (defaults to the signed-in member)")
	recordingRequestsCreateCmd.Flags().String(
		"request-instructions",
		"",
		"Instructions shown to the person who will upload the recording",
	)
	recordingRequestsCreateCmd.Flags().String("request-notes", "", "Private notes for this request")
	_ = recordingRequestsCreateCmd.MarkFlagRequired("title")

	recordingRequestsCmd.AddCommand(recordingRequestsCreateCmd)
	register(func(root *cobra.Command) { root.AddCommand(recordingRequestsCmd) })
}
