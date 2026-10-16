package commands

import "testing"

func TestTagsListSendsSortParams(t *testing.T) {
	req := runCommand(t, `{"tags":[],"pagination":{}}`,
		"tags", "list", "--sort-by", "created_at", "--sort-direction", "desc")

	if req.path != "/api/external/v2/tags" {
		t.Fatalf("path = %s", req.path)
	}
	if req.query.Get("sort_by") != "created_at" || req.query.Get("sort_direction") != "desc" {
		t.Errorf("query = %v", req.query)
	}
}
