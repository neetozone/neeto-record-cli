package commands

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestRecordingRequestsCreateCommandExists(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"recording-requests", "create"})
	if err != nil {
		t.Fatalf("recording-requests create command not found: %v", err)
	}

	if cmd.Use != "create" {
		t.Fatalf("unexpected command use: %q", cmd.Use)
	}
}

func TestRecordingRequestsCreateRequiredFlags(t *testing.T) {
	for _, name := range []string{"title", "created-by-email"} {
		flag := recordingRequestsCreateCmd.Flag(name)
		if flag == nil {
			t.Fatalf("flag %q not found", name)
		}
		if _, ok := flag.Annotations[cobra.BashCompOneRequiredFlag]; !ok {
			t.Fatalf("flag %q is not marked as required", name)
		}
	}
}
