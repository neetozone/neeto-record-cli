package commands

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"github.com/neetozone/neeto-cli-commons/auth"
	"github.com/neetozone/neeto-cli-commons/cli"
	"github.com/neetozone/neeto-cli-commons/config"
	product "github.com/neetozone/neeto-record-cli"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type capturedRequest struct {
	method string
	path   string
	query  url.Values
	body   string
}

func runCommand(t *testing.T, response string, args ...string) capturedRequest {
	t.Helper()
	t.Setenv("HOME", t.TempDir())

	cfg, err := config.Parse(product.ConfigYAML)
	if err != nil {
		t.Fatalf("config.Parse() error = %v", err)
	}
	testApp := cli.New(*cfg)
	Register(testApp)

	if err := testApp.Auth.SaveStore(&auth.Store{Credentials: []auth.Credentials{
		{Subdomain: "acme", Email: "dev@acme.com", SessionToken: "session-token"},
	}}); err != nil {
		t.Fatalf("SaveStore() error = %v", err)
	}

	var captured capturedRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		captured = capturedRequest{method: r.Method, path: r.URL.Path, query: r.URL.Query(), body: string(body)}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()
	t.Setenv("NEETORECORD_BASE_URL", server.URL)

	resetFlags(testApp.Root())
	testApp.Root().SetArgs(args)
	if err := testApp.Root().Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	return captured
}

func resetFlags(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if sv, ok := f.Value.(pflag.SliceValue); ok {
			_ = sv.Replace(nil)
		} else {
			_ = f.Value.Set(f.DefValue)
		}
		f.Changed = false
	})
	for _, sub := range cmd.Commands() {
		resetFlags(sub)
	}
}

func TestRecordingsListSendsFilters(t *testing.T) {
	req := runCommand(t, `{"recordings":[],"pagination":{}}`,
		"recordings", "list",
		"--owner-email", "me",
		"--from-date", "2026-10-01",
		"--to-date", "2026-10",
		"--folder-id", "fld_1",
		"--tag", "demo",
		"--tag", "sales",
		"--title", "onboarding",
		"--has-transcript=false",
		"--requested-only",
		"--sort-by", "duration",
		"--sort-direction", "asc",
	)

	want := url.Values{
		"owner_email":    {"me"},
		"from_date":      {"2026-10-01"},
		"to_date":        {"2026-10"},
		"folder_id":      {"fld_1"},
		"tag_names[]":    {"demo", "sales"},
		"title":          {"onboarding"},
		"has_transcript": {"false"},
		"requested_only": {"true"},
		"sort_by":        {"duration"},
		"sort_direction": {"asc"},
	}
	if req.method != http.MethodGet || req.path != "/api/external/v2/recordings" {
		t.Fatalf("request = %s %s", req.method, req.path)
	}
	if !reflect.DeepEqual(req.query, want) {
		t.Errorf("query = %v, want %v", req.query, want)
	}
}

func TestRecordingsListSendsNoFilterParamsWhenFlagsUnset(t *testing.T) {
	req := runCommand(t, `{"recordings":[],"pagination":{}}`, "recordings", "list")

	if len(req.query) != 0 {
		t.Errorf("query = %v, want empty", req.query)
	}
}

func TestRecordingsListSendsHasTranscriptTrue(t *testing.T) {
	req := runCommand(t, `{"recordings":[],"pagination":{}}`, "recordings", "list", "--has-transcript")

	if got := req.query.Get("has_transcript"); got != "true" {
		t.Errorf("has_transcript = %q, want true", got)
	}
}

func TestRecordingsActionItemsPath(t *testing.T) {
	req := runCommand(t, `{"recording_id":"abc","status":"success","action_items":[]}`,
		"recordings", "action-items", "abc")

	if req.method != http.MethodGet || req.path != "/api/external/v2/recordings/abc/action-items" {
		t.Fatalf("request = %s %s", req.method, req.path)
	}
}

func TestRecordingsTriggerActionItemsPath(t *testing.T) {
	req := runCommand(t, `{"recording_id":"abc","status":"in_progress","action_items":[]}`,
		"recordings", "trigger-action-items", "abc")

	if req.method != http.MethodPost || req.path != "/api/external/v2/recordings/abc/trigger-action-items" {
		t.Fatalf("request = %s %s", req.method, req.path)
	}
}

func TestRecordingsCtasSendsSortAndPagination(t *testing.T) {
	req := runCommand(t, `{"recording_id":"abc","ctas":[],"pagination":{}}`,
		"recordings", "ctas", "abc",
		"--page", "2", "--page-size", "10", "--sort-by", "label", "--sort-direction", "asc")

	want := url.Values{
		"page":           {"2"},
		"page_number":    {"2"},
		"page_size":      {"10"},
		"sort_by":        {"label"},
		"sort_direction": {"asc"},
	}
	if req.path != "/api/external/v2/recordings/abc/ctas" {
		t.Fatalf("path = %s", req.path)
	}
	if !reflect.DeepEqual(req.query, want) {
		t.Errorf("query = %v, want %v", req.query, want)
	}
}
