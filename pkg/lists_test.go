package pkg

import (
	"context"
	"encoding/json"
	"github.com/bluesky-social/indigo/api/bsky"
	"os"
	"testing"
	"time"
)

func Test_GetList(t *testing.T) {
	if os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skipf("Test_StarterPacks is a manual test that is skipped in CICD")
	}

	stuff, err := testSetup()
	if err != nil {
		t.Fatalf("testSetup() = %v, wanted nil", err)
	}

	// Lets test getting a list associated with a starter pack.
	// This is from Chris Albon's starter pack
	//listRef := "at://did:plc:umpsiyampiq3bpgce7kigydz/app.bsky.graph.list/3l7teemxtgy25"

	// Oneof the lists I created
	listRef := "at://did:plc:5lwweotr4gfb7bbz2fqwdthf/app.bsky.graph.list/3l7u3rez6hz2e"

	cursor := ""
	limit := int64(100)
	list, err := bsky.GraphGetList(context.Background(), stuff.Client, cursor, limit, listRef)

	if err != nil {
		t.Fatalf("GraphGetList() = %v, wanted nil", err)
	}

	b, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		t.Fatalf("json.Marshal() = %v, wanted nil", err)
	}
	t.Logf("list:\n%+s", string(b))
}

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
