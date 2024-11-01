package pkg

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
)

const (
	// My did for testing purposes
	jeremyLewiDid = "did:plc:5lwweotr4gfb7bbz2fqwdthf"
)

func Test_GetStarterPackOutput(t *testing.T) {
	// Verify the GoLang struct is correct by trying to deserialize some JSON
	// The JSON was obtained by making a client.Do request with a *bytes.Buffer for the output

	jsonString := `{"starterPacks":[{"uri":"at://did:plc:umpsiyampiq3bpgce7kigydz/app.bsky.graph.starterpack/3l7teencn4f2r","cid":"bafyreieahkxvemhdokzukhpn5bubbfhqcfs6jvpkpy5tfbtnutdnh2agge","record":{"$type":"app.bsky.graph.starterpack","createdAt":"2024-10-31T19:11:16.971Z","feeds":[],"list":"at://did:plc:umpsiyampiq3bpgce7kigydz/app.bsky.graph.list/3l7teemxtgy25","name":"Data Starter Pack Starter Pack"},"creator":{"did":"did:plc:umpsiyampiq3bpgce7kigydz","handle":"chrisalbon.com","displayName":"Chris Albon","avatar":"https://cdn.bsky.app/img/avatar/plain/did:plc:umpsiyampiq3bpgce7kigydz/bafkreiadswkm4njoz5fsp3irzxmhh6h2z6jmvpblh72zbp2uu7z2iexp3e@jpeg","associated":{"chat":{"allowIncoming":"all"}},"viewer":{"muted":false,"blockedBy":false,"following":"at://did:plc:5lwweotr4gfb7bbz2fqwdthf/app.bsky.graph.follow/3l7r3fg3shx2g","followedBy":"at://did:plc:umpsiyampiq3bpgce7kigydz/app.bsky.graph.follow/3l7l72e7da22a"},"labels":[],"createdAt":"2023-03-16T01:40:20.177Z"},"joinedAllTimeCount":0,"joinedWeekCount":0,"labels":[],"indexedAt":"2024-10-31T19:11:16.971Z"},{"uri":"at://did:plc:umpsiyampiq3bpgce7kigydz/app.bsky.graph.starterpack/3l7qpbmatnb2c","cid":"bafyreieyngelz5dwe3x5ktlbegqqaa7n5ni53j4xjbcltseko324gfsqqu","record":{"$type":"app.bsky.graph.starterpack","createdAt":"2024-10-30T17:48:27.214Z","list":"at://did:plc:umpsiyampiq3bpgce7kigydz/app.bsky.graph.list/3l7qpblunwu22","name":"Chris Albon's Starter Pack"},"creator":{"did":"did:plc:umpsiyampiq3bpgce7kigydz","handle":"chrisalbon.com","displayName":"Chris Albon","avatar":"https://cdn.bsky.app/img/avatar/plain/did:plc:umpsiyampiq3bpgce7kigydz/bafkreiadswkm4njoz5fsp3irzxmhh6h2z6jmvpblh72zbp2uu7z2iexp3e@jpeg","associated":{"chat":{"allowIncoming":"all"}},"viewer":{"muted":false,"blockedBy":false,"following":"at://did:plc:5lwweotr4gfb7bbz2fqwdthf/app.bsky.graph.follow/3l7r3fg3shx2g","followedBy":"at://did:plc:umpsiyampiq3bpgce7kigydz/app.bsky.graph.follow/3l7l72e7da22a"},"labels":[],"createdAt":"2023-03-16T01:40:20.177Z"},"joinedAllTimeCount":0,"joinedWeekCount":0,"labels":[],"indexedAt":"2024-10-30T17:48:27.214Z"}]}`

	b := bytes.NewBuffer([]byte(jsonString))
	var out GetStarterPacks_Output
	d := json.NewDecoder(b)
	d.DisallowUnknownFields()
	if err := d.Decode(&out); err != nil {
		t.Fatalf("json.Unmarshal() = %v, wanted nil", err)
	}
	t.Logf("out: %+v", out)
}

func Test_GetStarterPacks(t *testing.T) {
	if os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skipf("Test_StarterPacks is a manual test that is skipped in CICD")
	}

	stuff, err := testSetup()
	if err != nil {
		t.Fatalf("testSetup() = %v, wanted nil", err)
	}
	// SHould be Chris Albon who has a starter pack
	// actor := "did:plc:umpsiyampiq3bpgce7kigydz"
	actor := jeremyLewiDid

	out, err := GetStarterPacks(stuff.Client, actor)
	if err != nil {
		t.Fatalf("GetStarterPacks() = %v, wanted nil", err)
	}

	if len(out.StarterPacks) <= 0 {
		t.Fatalf("GetStarterPacks() = %v, wanted > 0", len(out.StarterPacks))
	}
}
