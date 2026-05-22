package commands

import (
	"github.com/neetozone/neeto-record-cli/internal/output"
	"github.com/spf13/cobra"
)

var tagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "List recording tags",
}

var tagsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tags",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		data, err := c.Get("/tags", paginationParams(cmd))
		if err != nil {
			return err
		}

		printList(data, "tags", []output.Breadcrumb{})
		return nil
	},
}

func init() {
	addPaginationFlags(tagsListCmd)

	tagsCmd.AddCommand(tagsListCmd)
	rootCmd.AddCommand(tagsCmd)
}
