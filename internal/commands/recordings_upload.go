package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/neetozone/neeto-cli-commons/output"
	"github.com/spf13/cobra"
)

const (
	maxUploadSize  = 2 << 30
	uploadAttempts = 3
)

var recordingsUploadCmd = &cobra.Command{
	Use:   "upload <file.mp4>",
	Short: "Upload an mp4 video as a recording",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]
		info, err := os.Stat(path)
		if err != nil || info.IsDir() {
			return fmt.Errorf("File %s does not exist", path)
		}
		if !strings.EqualFold(filepath.Ext(path), ".mp4") {
			return fmt.Errorf("File %s is not an .mp4 file", path)
		}
		if info.Size() > maxUploadSize {
			return fmt.Errorf("File %s is larger than the 2 GB upload limit", path)
		}

		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		body := map[string]interface{}{"file_size": info.Size()}
		if v, _ := cmd.Flags().GetString("title"); v != "" {
			body["title"] = v
		}
		if v, _ := cmd.Flags().GetString("folder-id"); v != "" {
			body["folder_id"] = v
		}

		data, err := c.Post("/recordings/uploads", body)
		if err != nil {
			return err
		}

		var created struct {
			Recording struct {
				ID string `json:"id"`
			} `json:"recording"`
			UploadURL     string            `json:"upload_url"`
			UploadHeaders map[string]string `json:"upload_headers"`
		}
		if err := json.Unmarshal(data, &created); err != nil {
			return fmt.Errorf("Could not read upload response: %w", err)
		}

		printer := app.Printer
		if !printer.Quiet && !printer.ForceJSON && !printer.Toon {
			spinner := printer.Spinner(fmt.Sprintf("Uploading %s", filepath.Base(path)))
			spinner.Start()
			defer spinner.Stop()
		}

		if err := putFile(path, info.Size(), created.UploadURL, created.UploadHeaders); err != nil {
			return err
		}

		data, err = c.Post(fmt.Sprintf("/recordings/%s/complete-upload", created.Recording.ID), nil)
		if err != nil {
			return err
		}

		printActionResult(data, []output.Breadcrumb{
			{Label: "Show", Command: "neetorecord recordings show <id>"},
		})
		return nil
	},
}

func putFile(path string, size int64, uploadURL string, headers map[string]string) error {
	var lastErr error
	for attempt := 0; attempt < uploadAttempts; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
		}
		retry, err := putFileOnce(path, size, uploadURL, headers)
		if err == nil {
			return nil
		}
		if !retry {
			return err
		}
		lastErr = err
	}
	return lastErr
}

func putFileOnce(path string, size int64, uploadURL string, headers map[string]string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer func() { _ = file.Close() }()

	req, err := http.NewRequest(http.MethodPut, uploadURL, file)
	if err != nil {
		return false, err
	}
	req.ContentLength = size
	for name, value := range headers {
		req.Header.Set(name, value)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return true, fmt.Errorf("Could not upload the file: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return resp.StatusCode >= 500, fmt.Errorf("Upload to storage failed with status %d", resp.StatusCode)
	}
	return false, nil
}

func init() {
	recordingsUploadCmd.Flags().String("title", "", "Title for the recording")
	recordingsUploadCmd.Flags().String("folder-id", "", "Folder ID to put the recording in")
	recordingsCmd.AddCommand(recordingsUploadCmd)
}
