package commands

import "testing"

func TestFoldersListSendsSortParams(t *testing.T) {
	req := runCommand(t, `{"folders":[],"pagination":{}}`,
		"folders", "list", "--sort-by", "name", "--sort-direction", "asc")

	if req.path != "/api/external/v2/folders" {
		t.Fatalf("path = %s", req.path)
	}
	if req.query.Get("sort_by") != "name" || req.query.Get("sort_direction") != "asc" {
		t.Errorf("query = %v", req.query)
	}
}

func TestFoldersListOmitsUnsetSortParams(t *testing.T) {
	req := runCommand(t, `{"folders":[],"pagination":{}}`, "folders", "list")

	if req.query.Has("sort_by") || req.query.Has("sort_direction") {
		t.Errorf("sort params should be omitted, query = %v", req.query)
	}
}
