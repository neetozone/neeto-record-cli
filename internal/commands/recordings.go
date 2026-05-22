package commands

import (
	"github.com/neetozone/neeto-record-cli/internal/output"
	"github.com/spf13/cobra"
)

var recordingsCmd = &cobra.Command{
	Use:   "recordings",
	Short: "Manage recordings",
}

var recordingsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List recordings",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		data, err := c.Get("/recordings", paginationParams(cmd))
		if err != nil {
			return err
		}

		printList(data, "recordings", []output.Breadcrumb{})
		return nil
	},
}

func init() {
	addPaginationFlags(recordingsListCmd)

	recordingsCmd.AddCommand(recordingsListCmd)
	rootCmd.AddCommand(recordingsCmd)
}
