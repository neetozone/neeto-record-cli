package commands

import (
	"github.com/neetozone/neeto-cli-commons/output"
	"github.com/spf13/cobra"
)

var foldersCmd = &cobra.Command{
	Use:   "folders",
	Short: "Manage recording folders",
}

var foldersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List folders",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		data, err := c.Get("/folders", paginationParams(cmd))
		if err != nil {
			return err
		}

		printList(data, "folders", []output.Breadcrumb{})
		return nil
	},
}

var foldersCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new folder",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		name, _ := cmd.Flags().GetString("name")
		folder := map[string]interface{}{"name": name}

		if v, _ := cmd.Flags().GetString("parent-folder-id"); cmd.Flags().Changed("parent-folder-id") {
			folder["parent_folder_id"] = v
		}

		data, err := c.Post("/folders", map[string]interface{}{"folder": folder})
		if err != nil {
			return err
		}

		printActionResult(data, nil)
		return nil
	},
}

func init() {
	addPaginationFlags(foldersListCmd)

	foldersCreateCmd.Flags().String("name", "", "Folder name")
	foldersCreateCmd.Flags().String("parent-folder-id", "", "Parent folder ID (for nested folders)")
	_ = foldersCreateCmd.MarkFlagRequired("name")

	foldersCmd.AddCommand(foldersListCmd)
	foldersCmd.AddCommand(foldersCreateCmd)
	register(func(root *cobra.Command) { root.AddCommand(foldersCmd) })
}
