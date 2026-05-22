package commands

import (
	"encoding/json"
	"fmt"

	"github.com/neetozone/neeto-record-cli/internal/output"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the CLI version",
	Run: func(cmd *cobra.Command, args []string) {
		if output.UseJSON() {
			payload, _ := json.Marshal(map[string]string{
				"binary":  "neetorecord",
				"version": Version,
				"commit":  Commit,
				"date":    Date,
			})
			fmt.Println(string(payload))
			return
		}
		fmt.Printf("neetorecord %s (commit: %s, built: %s)\n", Version, Commit, Date)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
