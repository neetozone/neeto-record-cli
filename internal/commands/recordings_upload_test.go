package commands

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neetozone/neeto-cli-commons/auth"
	"github.com/neetozone/neeto-cli-commons/cli"
	"github.com/neetozone/neeto-cli-commons/config"
	product "github.com/neetozone/neeto-record-cli"
)

func runUploadCommand(t *testing.T, api http.Handler, args ...string) (string, error) {
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

	server := httptest.NewServer(api)
	defer server.Close()
	t.Setenv("NEETORECORD_BASE_URL", server.URL)

	var out bytes.Buffer
	testApp.Printer.Out = &out
	resetFlags(testApp.Root())
	testApp.Root().SetArgs(args)
	err = testApp.Root().Execute()
	return out.String(), err
}

func writeFile(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	return path
}

func TestRecordingsUploadCreatesPutsAndCompletes(t *testing.T) {
	var putBody, putContentType, putAuth, createBody string
	var putLength int64
	var calls []string

	storage := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		calls = append(calls, r.Method+" storage")
		putBody, putContentType, putAuth, putLength = string(body), r.Header.Get("Content-Type"), r.Header.Get("Session-Token"), r.ContentLength
	}))
	defer storage.Close()

	api := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.URL.Path)
		body, _ := io.ReadAll(r.Body)
		switch r.URL.Path {
		case "/api/external/v2/recordings/uploads":
			createBody = string(body)
			_, _ = fmt.Fprintf(w, `{"recording":{"id":"rec_1","title":"Demo","url":"https://acme.neetorecord.com/watch/abc"},"upload_url":%q,"upload_headers":{"Content-Type":"video/mp4"}}`, storage.URL+"/put")
		case "/api/external/v2/recordings/rec_1/complete-upload":
			_, _ = w.Write([]byte(`{"recording":{"id":"rec_1","title":"Demo","url":"https://acme.neetorecord.com/watch/abc"}}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})

	path := writeFile(t, "demo.mp4", "video-bytes")
	out, err := runUploadCommand(t, api, "recordings", "upload", path, "--title", "Demo", "--folder-id", "fld_1", "--json")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	wantCalls := "POST /api/external/v2/recordings/uploads|PUT storage|POST /api/external/v2/recordings/rec_1/complete-upload"
	if got := strings.Join(calls, "|"); got != wantCalls {
		t.Errorf("calls = %q, want %q", got, wantCalls)
	}
	if want := `{"file_size":11,"folder_id":"fld_1","title":"Demo"}`; strings.TrimSpace(createBody) != want {
		t.Errorf("create body = %q, want %q", createBody, want)
	}
	if putBody != "video-bytes" || putLength != 11 {
		t.Errorf("PUT body = %q length %d, want video-bytes length 11", putBody, putLength)
	}
	if putContentType != "video/mp4" {
		t.Errorf("PUT Content-Type = %q, want video/mp4", putContentType)
	}
	if putAuth != "" {
		t.Errorf("PUT carried Session-Token %q, want none", putAuth)
	}
	if !strings.Contains(out, `"url": "https://acme.neetorecord.com/watch/abc"`) {
		t.Errorf("output = %q, want the watch url", out)
	}
}

func TestRecordingsUploadQuietPrintsOnlyTheIdentifier(t *testing.T) {
	storage := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer storage.Close()

	api := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/uploads") {
			_, _ = fmt.Fprintf(w, `{"recording":{"id":"rec_1"},"upload_url":%q,"upload_headers":{}}`, storage.URL)
			return
		}
		_, _ = w.Write([]byte(`{"recording":{"id":"rec_1","title":"Demo","url":"https://acme.neetorecord.com/watch/abc"}}`))
	})

	out, err := runUploadCommand(t, api, "recordings", "upload", writeFile(t, "demo.mp4", "x"), "--quiet")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if strings.TrimSpace(out) != "rec_1" {
		t.Errorf("output = %q, want rec_1", out)
	}
}

func TestRecordingsUploadRetriesAServerErrorFromStorage(t *testing.T) {
	puts := 0
	storage := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		puts++
		if puts == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
	}))
	defer storage.Close()

	api := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/uploads") {
			_, _ = fmt.Fprintf(w, `{"recording":{"id":"rec_1"},"upload_url":%q,"upload_headers":{}}`, storage.URL)
			return
		}
		_, _ = w.Write([]byte(`{"recording":{"id":"rec_1"}}`))
	})

	if _, err := runUploadCommand(t, api, "recordings", "upload", writeFile(t, "demo.mp4", "x"), "--quiet"); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if puts != 2 {
		t.Errorf("PUT attempts = %d, want 2", puts)
	}
}

func TestRecordingsUploadRetriesWhenStorageSaysTooManyRequests(t *testing.T) {
	puts := 0
	storage := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		puts++
		if puts == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
		}
	}))
	defer storage.Close()

	api := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/uploads") {
			_, _ = fmt.Fprintf(w, `{"recording":{"id":"rec_1"},"upload_url":%q,"upload_headers":{}}`, storage.URL)
			return
		}
		_, _ = w.Write([]byte(`{"recording":{"id":"rec_1"}}`))
	})

	if _, err := runUploadCommand(t, api, "recordings", "upload", writeFile(t, "demo.mp4", "x"), "--quiet"); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if puts != 2 {
		t.Errorf("PUT attempts = %d, want 2", puts)
	}
}

func TestRecordingsUploadDoesNotRetryAClientErrorFromStorage(t *testing.T) {
	puts := 0
	storage := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		puts++
		w.WriteHeader(http.StatusForbidden)
	}))
	defer storage.Close()

	api := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, `{"recording":{"id":"rec_1"},"upload_url":%q,"upload_headers":{}}`, storage.URL)
	})

	_, err := runUploadCommand(t, api, "recordings", "upload", writeFile(t, "demo.mp4", "x"), "--quiet")
	if err == nil || !strings.Contains(err.Error(), "403") {
		t.Fatalf("error = %v, want one naming status 403", err)
	}
	if puts != 1 {
		t.Errorf("PUT attempts = %d, want 1", puts)
	}
}

func TestRecordingsUploadRefusesFilesLocally(t *testing.T) {
	api := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("unexpected API call %s", r.URL.Path)
	})

	big := writeFile(t, "big.mp4", "")
	if err := os.Truncate(big, 2<<30+1); err != nil {
		t.Fatalf("Truncate() error = %v", err)
	}

	tests := map[string]struct {
		path string
		want string
	}{
		"missing file": {filepath.Join(t.TempDir(), "nope.mp4"), "does not exist"},
		"not an mp4":   {writeFile(t, "clip.mov", "x"), "not an .mp4"},
		"over 2 GB":    {big, "2 GB"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := runUploadCommand(t, api, "recordings", "upload", tc.path)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error = %v, want one containing %q", err, tc.want)
			}
		})
	}
}

func TestRecordingsUploadShowsTheServerMessageWhenCreateIsForbidden(t *testing.T) {
	api := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"Your plan does not allow uploads"}`))
	})

	_, err := runUploadCommand(t, api, "recordings", "upload", writeFile(t, "demo.mp4", "x"))
	if err == nil || !strings.Contains(err.Error(), "Your plan does not allow uploads") {
		t.Errorf("error = %v, want the server message", err)
	}
}
