package gcp

import (
	"log"
	"testing"
)

func TestIdentify(t *testing.T) {
	test_project_id := "test-project-id"
	gcp := GCP{
		ProjectID: test_project_id,
	}

	identification := gcp.Identify()
	log.Println(identification)

	if identification["project_id"] != test_project_id {
		t.Error("TestIdentify failed")
	}
}
