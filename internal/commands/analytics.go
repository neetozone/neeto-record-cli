package commands

import (
	"net/url"

	"github.com/spf13/cobra"
)

var analyticsCmd = &cobra.Command{
	Use:   "analytics",
	Short: "View organization analytics",
}

var analyticsShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show organization-wide analytics",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		params := url.Values{}
		if v, _ := cmd.Flags().GetString("from-date"); cmd.Flags().Changed("from-date") {
			params.Set("from_date", v)
		}
		if v, _ := cmd.Flags().GetString("to-date"); cmd.Flags().Changed("to-date") {
			params.Set("to_date", v)
		}

		data, err := c.Get("/analytics", params)
		if err != nil {
			return err
		}

		printResource(data, nil)
		return nil
	},
}

func init() {
	analyticsShowCmd.Flags().String("from-date", "", "Start date filter (ISO format, e.g. '2024-01-01')")
	analyticsShowCmd.Flags().String("to-date", "", "End date filter (ISO format, e.g. '2024-12-31')")

	analyticsCmd.AddCommand(analyticsShowCmd)
	rootCmd.AddCommand(analyticsCmd)
}
