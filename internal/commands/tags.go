package commands

import (
	"github.com/neetozone/neeto-cli-commons/output"
	"github.com/spf13/cobra"
)

var tagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "List recording tags",
}

var tagsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tags",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		params := paginationParams(cmd)
		setChangedFlags(cmd, params, "sort-by", "sort-direction")

		data, err := c.Get("/tags", params)
		if err != nil {
			return err
		}

		printList(data, "tags", []output.Breadcrumb{})
		return nil
	},
}

func init() {
	addPaginationFlags(tagsListCmd)
	addSortFlags(tagsListCmd, "created_at or name")

	tagsCmd.AddCommand(tagsListCmd)
	register(func(root *cobra.Command) { root.AddCommand(tagsCmd) })
}
