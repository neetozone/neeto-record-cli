package output

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("io.Copy error: %v", err)
	}
	return buf.String()
}

func TestUseJSON_ForceJSON(t *testing.T) {
	ForceJSON = true
	QuietMode = false
	defer func() { ForceJSON = false }()

	if !UseJSON() {
		t.Error("UseJSON() = false, want true when ForceJSON is set")
	}
}

func TestUseJSON_QuietMode(t *testing.T) {
	ForceJSON = false
	QuietMode = true
	defer func() { QuietMode = false }()

	if !UseJSON() {
		t.Error("UseJSON() = false, want true when QuietMode is set")
	}
}

func TestPrintMessage_JSON(t *testing.T) {
	ForceJSON = true
	QuietMode = false
	defer func() { ForceJSON = false }()

	out := captureStdout(t, func() {
		PrintMessage("hello world")
	})

	var parsed map[string]string
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if parsed["message"] != "hello world" {
		t.Errorf("message = %q, want %q", parsed["message"], "hello world")
	}
}

func TestPrintMessage_Plain(t *testing.T) {
	ForceJSON = false
	QuietMode = false

	out := captureStdout(t, func() {
		PrintMessage("hello world")
	})

	trimmed := strings.TrimSpace(out)
	if !strings.Contains(trimmed, "hello world") {
		t.Errorf("output = %q, want it to contain %q", trimmed, "hello world")
	}
}

func TestPrint_QuietMode(t *testing.T) {
	ForceJSON = false
	QuietMode = true
	defer func() { QuietMode = false }()

	data := json.RawMessage(`[{"id":1}]`)
	out := captureStdout(t, func() {
		Print(data, nil)
	})

	trimmed := strings.TrimSpace(out)
	if trimmed != `[{"id":1}]` {
		t.Errorf("quiet output = %q, want raw data", trimmed)
	}
}

func TestPrint_JSONEnvelope(t *testing.T) {
	ForceJSON = true
	QuietMode = false
	defer func() { ForceJSON = false }()

	data := json.RawMessage(`{"name":"test"}`)
	breadcrumbs := []Breadcrumb{{Label: "details", Command: "app show 1"}}

	out := captureStdout(t, func() {
		Print(data, breadcrumbs)
	})

	var envelope Envelope
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &envelope); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	var parsed map[string]string
	if err := json.Unmarshal(envelope.Data, &parsed); err != nil {
		t.Fatalf("envelope data is not valid JSON: %v", err)
	}
	if parsed["name"] != "test" {
		t.Errorf("data.name = %q, want %q", parsed["name"], "test")
	}
	if len(envelope.Breadcrumbs) != 1 || envelope.Breadcrumbs[0].Label != "details" {
		t.Errorf("breadcrumbs = %v, want [{details app show 1}]", envelope.Breadcrumbs)
	}
}

func TestPrintWithPagination_QuietMode(t *testing.T) {
	ForceJSON = false
	QuietMode = true
	defer func() { QuietMode = false }()

	data := json.RawMessage(`[{"id":1}]`)
	pagination := json.RawMessage(`{"current_page_number":1,"total_pages":3,"total_records":25}`)

	out := captureStdout(t, func() {
		PrintWithPagination(data, pagination, nil)
	})

	trimmed := strings.TrimSpace(out)
	if trimmed != `[{"id":1}]` {
		t.Errorf("quiet output = %q, want raw data without pagination", trimmed)
	}
}

func TestPrintWithPagination_JSONEnvelope(t *testing.T) {
	ForceJSON = true
	QuietMode = false
	defer func() { ForceJSON = false }()

	data := json.RawMessage(`[{"id":1}]`)
	pagination := json.RawMessage(`{"current_page_number":1,"total_pages":3}`)

	out := captureStdout(t, func() {
		PrintWithPagination(data, pagination, nil)
	})

	var envelope Envelope
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &envelope); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if envelope.Pagination == nil {
		t.Error("pagination should be present in envelope")
	}
}

func TestPickColumns_PrioritisesNeetoRecordFields(t *testing.T) {
	tests := []struct {
		name   string
		sample map[string]interface{}
		want   []string
	}{
		{
			name: "recordings",
			sample: map[string]interface{}{
				"id": "", "title": "", "default_title": "", "duration": 1.0,
				"view_count": 1.0, "is_uploaded": true, "requested": false,
				"summary": "", "public_link_id": "", "created_at": "", "updated_at": "",
				"uploaded_at": "", "public_url": "", "transcoded_url": "",
				"thumbnail_url": "", "user_name": "", "folder_name": "", "folder_id": "",
			},
			want: []string{"id", "title", "duration", "view_count", "user_name", "folder_name", "created_at"},
		},
		{
			name: "team members",
			sample: map[string]interface{}{
				"id": "", "email": "", "first_name": "", "last_name": "",
				"time_zone": "", "profile_image_url": nil, "active": true,
				"organization_role": "",
			},
			want: []string{"id", "email", "first_name", "last_name", "organization_role", "time_zone", "active"},
		},
		{
			name: "folders",
			sample: map[string]interface{}{
				"id": "", "name": "", "parent_id": nil, "recording_count": 1.0, "created_at": "",
			},
			want: []string{"id", "name", "recording_count", "created_at", "parent_id"},
		},
		{
			name: "tags",
			sample: map[string]interface{}{
				"id": "", "name": "", "style": "", "recording_count": 1.0,
			},
			want: []string{"id", "name", "recording_count", "style"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pickColumns(tt.sample)
			if len(got) != len(tt.want) {
				t.Fatalf("pickColumns() = %v, want %v", got, tt.want)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("pickColumns()[%d] = %q, want %q (full: %v)", i, got[i], tt.want[i], got)
				}
			}
		})
	}
}

// created_at and active must hold their columns because they are listed in
// priorityFields, not because they happen to sort first among the remaining
// fields. A field that sorts earlier must not displace them.
func TestPickColumns_PriorityFieldsBeatAlphabeticalFallback(t *testing.T) {
	recording := map[string]interface{}{
		"account_id": "", "id": "", "title": "", "duration": 1.0, "view_count": 1.0,
		"user_name": "", "folder_name": "", "created_at": "", "updated_at": "",
	}
	got := pickColumns(recording)
	if got[len(got)-1] != "created_at" {
		t.Errorf("created_at was displaced by an alphabetically earlier field: %v", got)
	}

	teamMember := map[string]interface{}{
		"access_level": "", "id": "", "email": "", "first_name": "", "last_name": "",
		"organization_role": "", "time_zone": "", "active": true,
	}
	got = pickColumns(teamMember)
	if got[len(got)-1] != "active" {
		t.Errorf("active was displaced by an alphabetically earlier field: %v", got)
	}
}
