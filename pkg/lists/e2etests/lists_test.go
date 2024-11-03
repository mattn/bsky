package e2etests

import (
	"context"
	"encoding/json"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/jlewi/bsctl/pkg/lists"
	"github.com/jlewi/bsctl/pkg/testutil"
	"os"
	"testing"
)

const (
	chrisAlbonDid = "did:plc:umpsiyampiq3bpgce7kigydz"
)

// N.B. Manual tests need to be in a subpackage to avoid circular dependencies because we reuse the app class.

func Test_GetList(t *testing.T) {
	if os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skipf("Test_StarterPacks is a manual test that is skipped in CICD")
	}

	stuff, err := testutil.New()
	if err != nil {
		t.Fatalf("testSetup() = %v, wanted nil", err)
	}

	// Lets test getting a list associated with a starter pack.
	// This is from Chris Albon's starter pack
	//listRef := "at://did:plc:umpsiyampiq3bpgce7kigydz/app.bsky.graph.list/3l7teemxtgy25"

	// Oneof the lists I created
	//listRef := "at://did:plc:5lwweotr4gfb7bbz2fqwdthf/app.bsky.graph.list/3l7u3rez6hz2e"

	// This is the list associated with my platform engineering starter pack
	listRef := "at://did:plc:5lwweotr4gfb7bbz2fqwdthf/app.bsky.graph.list/3l7u5daz2qa2w"

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

func Test_AddToStarterPackList(t *testing.T) {
	if os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skipf("Test_StarterPacks is a manual test that is skipped in CICD")
	}

	stuff, err := testutil.New()
	if err != nil {
		t.Fatalf("testSetup() = %v, wanted nil", err)
	}

	// This is the list associated with my platform engineering starter pack
	listRef := "at://did:plc:5lwweotr4gfb7bbz2fqwdthf/app.bsky.graph.list/3l7u5daz2qa2w"

	if err := lists.AddToList(stuff.Client, listRef, chrisAlbonDid); err != nil {
		t.Fatalf("AddToList returned error: %+v", err)
	}
}

func Test_CreateList(t *testing.T) {
	if os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skipf("Test_StarterPacks is a manual test that is skipped in CICD")
	}

	stuff, err := testutil.New()
	if err != nil {
		t.Fatalf("testSetup() = %v, wanted nil", err)
	}
	// SHould be Chris Albon
	//actor := "did:plc:umpsiyampiq3bpgce7kigydz"

	// Try setting it to the host for my PDS
	// that didn't work
	//stuff.Client.Host = "https://morel.us-east.host.bsky.network"
	_, err = lists.CreateListRecord(stuff.Client, "AI Engineering Community", "List of members of the AIEngineering community. Used for the feed")
	if err != nil {
		t.Fatalf("CreateListRecord() = %v, wanted nil", err)
	}
}
