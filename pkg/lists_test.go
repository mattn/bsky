package pkg

import (
	"os"
	"testing"
	"time"
)

func Test_CreateList(t *testing.T) {
	if os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skipf("Test_StarterPacks is a manual test that is skipped in CICD")
	}

	stuff, err := testSetup()
	if err != nil {
		t.Fatalf("testSetup() = %v, wanted nil", err)
	}
	// SHould be Chris Albon
	//actor := "did:plc:umpsiyampiq3bpgce7kigydz"

	listRecord := &ListRecord{
		Type:        "app.bsky.graph.list",
		CreatedAt:   time.Now().UTC(),
		Name:        "Test programmatically creating a list",
		Description: "A list of developers you might want to follow",
		Users: []User{
			{DID: "did:example:123"},
			{DID: "did:example:456"},
			{DID: "did:example:789"},
		},
	}

	// Try setting it to the host for my PDS
	// that didn't work
	//stuff.Client.Host = "https://morel.us-east.host.bsky.network"
	err = CreateListRecord(stuff.Client, listRecord)
	if err != nil {
		t.Fatalf("CreateListRecord() = %v, wanted nil", err)
	}
}
