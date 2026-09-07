package commands

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/neetozone/neeto-cli-commons/output"
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

		printList(data, "recordings", []output.Breadcrumb{
			{Label: "Show", Command: "neetorecord recordings show <id>"},
		})
		return nil
	},
}

var recordingsShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show a recording",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		data, err := c.Get(fmt.Sprintf("/recordings/%s", args[0]), nil)
		if err != nil {
			return err
		}

		printResource(data, nil)
		return nil
	},
}

var recordingsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a recording",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		recording := map[string]interface{}{}
		if v, _ := cmd.Flags().GetString("title"); cmd.Flags().Changed("title") {
			recording["title"] = v
		}
		if v, _ := cmd.Flags().GetString("summary"); cmd.Flags().Changed("summary") {
			recording["summary"] = v
		}
		if v, _ := cmd.Flags().GetString("folder-id"); cmd.Flags().Changed("folder-id") {
			recording["folder_id"] = v
		}
		if v, _ := cmd.Flags().GetStringArray("tag"); cmd.Flags().Changed("tag") {
			recording["tag_names"] = v
		}

		data, err := c.Patch(fmt.Sprintf("/recordings/%s", args[0]), map[string]interface{}{"recording": recording})
		if err != nil {
			return err
		}

		printResource(data, nil)
		return nil
	},
}

var recordingsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a recording",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := c.Delete(fmt.Sprintf("/recordings/%s", args[0])); err != nil {
			return err
		}

		printMessage("Recording deleted.")
		return nil
	},
}

var recordingsSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search recordings by title",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		query, _ := cmd.Flags().GetString("query")
		params := paginationParams(cmd)
		params.Set("query", query)

		data, err := c.Get("/recordings/search", params)
		if err != nil {
			return err
		}

		printList(data, "recordings", []output.Breadcrumb{
			{Label: "Show", Command: "neetorecord recordings show <id>"},
		})
		return nil
	},
}

var recordingsSearchByTranscriptCmd = &cobra.Command{
	Use:   "search-by-transcript",
	Short: "Search recordings by transcript content",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		query, _ := cmd.Flags().GetString("query")
		params := paginationParams(cmd)
		params.Set("query", query)

		data, err := c.Get("/recordings/search-by-transcript", params)
		if err != nil {
			return err
		}

		printList(data, "recordings", []output.Breadcrumb{
			{Label: "Show", Command: "neetorecord recordings show <id>"},
		})
		return nil
	},
}

var recordingsChaptersCmd = &cobra.Command{
	Use:   "chapters <id>",
	Short: "Get chapters for a recording",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		data, err := c.Get(fmt.Sprintf("/recordings/%s/chapters", args[0]), nil)
		if err != nil {
			return err
		}

		printResource(data, nil)
		return nil
	},
}

var recordingsTranscriptCmd = &cobra.Command{
	Use:   "transcript <id>",
	Short: "Get transcript for a recording",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		data, err := c.Get(fmt.Sprintf("/recordings/%s/transcript", args[0]), nil)
		if err != nil {
			return err
		}

		printResource(data, nil)
		return nil
	},
}

var recordingsShareLinkCmd = &cobra.Command{
	Use:   "share-link <id>",
	Short: "Get the public share link for a recording",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		data, err := c.Get(fmt.Sprintf("/recordings/%s/share-link", args[0]), nil)
		if err != nil {
			return err
		}

		printResource(data, nil)
		return nil
	},
}

var recordingsEmbedCodeCmd = &cobra.Command{
	Use:   "embed-code <id>",
	Short: "Get the embed code for a recording",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		data, err := c.Get(fmt.Sprintf("/recordings/%s/embed-code", args[0]), nil)
		if err != nil {
			return err
		}

		printResource(data, nil)
		return nil
	},
}

var recordingsDownloadUrlCmd = &cobra.Command{
	Use:   "download-url <id>",
	Short: "Get a presigned download URL for a recording",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		params := make(map[string]string)
		if format, _ := cmd.Flags().GetString("format"); format != "" {
			params["file_format"] = format
		}

		urlParams := paginationParams(cmd)
		for k, v := range params {
			urlParams.Set(k, v)
		}

		data, err := c.Get(fmt.Sprintf("/recordings/%s/download-url", args[0]), urlParams)
		if err != nil {
			return err
		}

		printResource(data, nil)
		return nil
	},
}

var recordingsScreenshotCmd = &cobra.Command{
	Use:   "screenshot <id>",
	Short: "Get a URL to a still-frame screenshot of a recording at a timestamp",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		timestamp, _ := cmd.Flags().GetFloat64("timestamp")
		format, _ := cmd.Flags().GetString("format")

		params := url.Values{}
		params.Set("timestamp", strconv.FormatFloat(timestamp, 'f', -1, 64))
		params.Set("image_format", format)

		data, err := c.Get(fmt.Sprintf("/recordings/%s/screenshot", args[0]), params)
		if err != nil {
			return err
		}

		printResource(data, nil)
		return nil
	},
}

var recordingsChapterStatusCmd = &cobra.Command{
	Use:   "chapter-status <id>",
	Short: "Check chapter generation status for a recording",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		data, err := c.Get(fmt.Sprintf("/recordings/%s/chapter-status", args[0]), nil)
		if err != nil {
			return err
		}

		printResource(data, nil)
		return nil
	},
}

var recordingsTranscriptStatusCmd = &cobra.Command{
	Use:   "transcript-status <id>",
	Short: "Check transcript generation status for a recording",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		data, err := c.Get(fmt.Sprintf("/recordings/%s/transcript-status", args[0]), nil)
		if err != nil {
			return err
		}

		printResource(data, nil)
		return nil
	},
}

var recordingsAnalyticsCmd = &cobra.Command{
	Use:   "analytics <id>",
	Short: "Get view analytics for a recording",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		data, err := c.Get(fmt.Sprintf("/recordings/%s/analytics", args[0]), nil)
		if err != nil {
			return err
		}

		printResource(data, nil)
		return nil
	},
}

var recordingsCtasCmd = &cobra.Command{
	Use:   "ctas <id>",
	Short: "List CTAs for a recording",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		data, err := c.Get(fmt.Sprintf("/recordings/%s/ctas", args[0]), nil)
		if err != nil {
			return err
		}

		printResource(data, nil)
		return nil
	},
}

var recordingsCreateCtaCmd = &cobra.Command{
	Use:   "create-cta <id>",
	Short: "Add a CTA to a recording",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		label, _ := cmd.Flags().GetString("label")
		start, _ := cmd.Flags().GetFloat64("start")
		end, _ := cmd.Flags().GetFloat64("end")

		cta := map[string]interface{}{
			"label": label,
			"start": start,
			"end":   end,
		}

		if v, _ := cmd.Flags().GetString("link"); cmd.Flags().Changed("link") {
			cta["link"] = v
		}
		if v, _ := cmd.Flags().GetString("position"); cmd.Flags().Changed("position") {
			cta["position"] = v
		}
		if v, _ := cmd.Flags().GetString("background-color"); cmd.Flags().Changed("background-color") {
			cta["background_color"] = v
		}
		if v, _ := cmd.Flags().GetString("text-color"); cmd.Flags().Changed("text-color") {
			cta["text_color"] = v
		}
		if v, _ := cmd.Flags().GetBool("show-only-at-end"); cmd.Flags().Changed("show-only-at-end") {
			cta["show_only_at_end"] = v
		}

		data, err := c.Post(fmt.Sprintf("/recordings/%s/ctas", args[0]), map[string]interface{}{"cta": cta})
		if err != nil {
			return err
		}

		printActionResult(data, nil)
		return nil
	},
}

var recordingsTriggerTranscriptCmd = &cobra.Command{
	Use:   "trigger-transcript <id>",
	Short: "Trigger transcript generation for a recording",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		body := map[string]interface{}{}
		if v, _ := cmd.Flags().GetString("language"); cmd.Flags().Changed("language") {
			body["language"] = v
		}

		data, err := c.Post(fmt.Sprintf("/recordings/%s/trigger-transcript", args[0]), body)
		if err != nil {
			return err
		}

		printActionResult(data, nil)
		return nil
	},
}

var recordingsTriggerChaptersCmd = &cobra.Command{
	Use:   "trigger-chapters <id>",
	Short: "Trigger chapter generation for a recording",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		data, err := c.Post(fmt.Sprintf("/recordings/%s/trigger-chapters", args[0]), nil)
		if err != nil {
			return err
		}

		printActionResult(data, nil)
		return nil
	},
}

var recordingsTriggerMp4Cmd = &cobra.Command{
	Use:   "trigger-mp4 <id>",
	Short: "Trigger MP4 generation for a recording",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		data, err := c.Post(fmt.Sprintf("/recordings/%s/trigger-mp4", args[0]), nil)
		if err != nil {
			return err
		}

		printActionResult(data, nil)
		return nil
	},
}

func init() {
	addPaginationFlags(recordingsListCmd)

	recordingsUpdateCmd.Flags().String("title", "", "New title")
	recordingsUpdateCmd.Flags().String("summary", "", "New summary")
	recordingsUpdateCmd.Flags().String("folder-id", "", "Folder ID to move to (empty to remove from folder)")
	recordingsUpdateCmd.Flags().StringArray("tag", nil, "Tag name (repeatable; replaces all existing tags)")

	addPaginationFlags(recordingsSearchCmd)
	recordingsSearchCmd.Flags().String("query", "", "Title search query")
	_ = recordingsSearchCmd.MarkFlagRequired("query")

	addPaginationFlags(recordingsSearchByTranscriptCmd)
	recordingsSearchByTranscriptCmd.Flags().String("query", "", "Text to search in transcripts")
	_ = recordingsSearchByTranscriptCmd.MarkFlagRequired("query")

	recordingsDownloadUrlCmd.Flags().String("format", "mp4", "Download format: 'mp4' or 'webm'")

	recordingsScreenshotCmd.Flags().Float64("timestamp", 0, "Timestamp in seconds of the frame to capture")
	recordingsScreenshotCmd.Flags().String("format", "png", "Image format: 'png' or 'jpeg'")
	_ = recordingsScreenshotCmd.MarkFlagRequired("timestamp")

	recordingsCreateCtaCmd.Flags().String("label", "", "Button label text")
	recordingsCreateCtaCmd.Flags().Float64("start", 0, "Start time in seconds")
	recordingsCreateCtaCmd.Flags().Float64("end", 0, "End time in seconds")
	recordingsCreateCtaCmd.Flags().String("link", "", "URL the button links to")
	recordingsCreateCtaCmd.Flags().String("position", "top-right", "Position: top-right, top-left, bottom-right, bottom-left")
	recordingsCreateCtaCmd.Flags().String("background-color", "#000000", "Button background color hex")
	recordingsCreateCtaCmd.Flags().String("text-color", "#FFFFFF", "Button text color hex")
	recordingsCreateCtaCmd.Flags().Bool("show-only-at-end", false, "Only show at end of recording")
	_ = recordingsCreateCtaCmd.MarkFlagRequired("label")
	_ = recordingsCreateCtaCmd.MarkFlagRequired("start")
	_ = recordingsCreateCtaCmd.MarkFlagRequired("end")

	recordingsTriggerTranscriptCmd.Flags().String("language", "", "Language code (e.g. 'en')")

	recordingsCmd.AddCommand(recordingsListCmd)
	recordingsCmd.AddCommand(recordingsShowCmd)
	recordingsCmd.AddCommand(recordingsUpdateCmd)
	recordingsCmd.AddCommand(recordingsDeleteCmd)
	recordingsCmd.AddCommand(recordingsSearchCmd)
	recordingsCmd.AddCommand(recordingsSearchByTranscriptCmd)
	recordingsCmd.AddCommand(recordingsChaptersCmd)
	recordingsCmd.AddCommand(recordingsTranscriptCmd)
	recordingsCmd.AddCommand(recordingsShareLinkCmd)
	recordingsCmd.AddCommand(recordingsEmbedCodeCmd)
	recordingsCmd.AddCommand(recordingsDownloadUrlCmd)
	recordingsCmd.AddCommand(recordingsScreenshotCmd)
	recordingsCmd.AddCommand(recordingsChapterStatusCmd)
	recordingsCmd.AddCommand(recordingsTranscriptStatusCmd)
	recordingsCmd.AddCommand(recordingsAnalyticsCmd)
	recordingsCmd.AddCommand(recordingsCtasCmd)
	recordingsCmd.AddCommand(recordingsCreateCtaCmd)
	recordingsCmd.AddCommand(recordingsTriggerTranscriptCmd)
	recordingsCmd.AddCommand(recordingsTriggerChaptersCmd)
	recordingsCmd.AddCommand(recordingsTriggerMp4Cmd)
	register(func(root *cobra.Command) { root.AddCommand(recordingsCmd) })
}
