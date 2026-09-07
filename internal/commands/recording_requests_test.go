package commands

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/neetozone/neeto-cli-commons/auth"
	"github.com/neetozone/neeto-cli-commons/cli"
	"github.com/neetozone/neeto-cli-commons/config"
	product "github.com/neetozone/neeto-record-cli"
	"github.com/spf13/cobra"
)

func TestRecordingRequestsCreateCommandExists(t *testing.T) {
	cmd, _, err := recordingRequestsCmd.Find([]string{"create"})
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

func TestRecordingRequestsCreatePayloadOmitsUnsetOptionals(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	cfg, err := config.Parse(product.ConfigYAML)
	if err != nil {
		t.Fatalf("config.Parse() error = %v", err)
	}
	testApp := cli.New(*cfg)
	Register(testApp)

	if err := testApp.Auth.SaveStore(&auth.Store{Credentials: []auth.Credentials{
		{
			Subdomain:    "acme",
			Email:        "dev@acme.com",
			SessionToken: "session-token",
		},
	}}); err != nil {
		t.Fatalf("SaveStore() error = %v", err)
	}

	var postedBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want %q", r.Method, http.MethodPost)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/api/external/v2/recording_requests" {
			t.Errorf("path = %q, want %q", r.URL.Path, "/api/external/v2/recording_requests")
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if err := json.NewDecoder(r.Body).Decode(&postedBody); err != nil {
			t.Errorf("Decode() error = %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"recording_request":{"id":"rq_123"}}`))
	}))
	defer server.Close()

	t.Setenv("NEETORECORD_BASE_URL", server.URL)

	testApp.Root().SetArgs([]string{
		"recording-requests", "create",
		"--title", "Q2 Demo Request",
		"--created-by-email", "requester@example.com",
	})
	if err := testApp.Root().Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	recordingPayload, ok := postedBody["recording"].(map[string]interface{})
	if !ok {
		t.Fatalf("request body missing recording wrapper, got: %#v", postedBody)
	}

	if recordingPayload["title"] != "Q2 Demo Request" {
		t.Fatalf("recording.title = %v, want %q", recordingPayload["title"], "Q2 Demo Request")
	}
	if recordingPayload["created_by_email"] != "requester@example.com" {
		t.Fatalf(
			"recording.created_by_email = %v, want %q",
			recordingPayload["created_by_email"],
			"requester@example.com",
		)
	}

	if _, ok := recordingPayload["request_instructions"]; ok {
		t.Fatalf("recording.request_instructions should be omitted when flag is unset")
	}
	if _, ok := recordingPayload["request_notes"]; ok {
		t.Fatalf("recording.request_notes should be omitted when flag is unset")
	}
}
