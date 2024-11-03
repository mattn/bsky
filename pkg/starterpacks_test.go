package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/jlewi/monogo/helpers"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"os"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"testing"
)

const (
	// My did for testing purposes
	jeremyLewiDid = "did:plc:5lwweotr4gfb7bbz2fqwdthf"
)

func Test_GetStarterPackOutput(t *testing.T) {
	// Verify the GoLang struct is correct by trying to deserialize some JSON
	// The JSON was obtained by making a client.Do request with a *bytes.Buffer for the output

	//jsonString := `{"starterPacks":[{"uri":"at://did:plc:umpsiyampiq3bpgce7kigydz/app.bsky.graph.starterpack/3l7teencn4f2r","cid":"bafyreieahkxvemhdokzukhpn5bubbfhqcfs6jvpkpy5tfbtnutdnh2agge","record":{"$type":"app.bsky.graph.starterpack","createdAt":"2024-10-31T19:11:16.971Z","feeds":[],"list":"at://did:plc:umpsiyampiq3bpgce7kigydz/app.bsky.graph.list/3l7teemxtgy25","name":"Data Starter Pack Starter Pack"},"creator":{"did":"did:plc:umpsiyampiq3bpgce7kigydz","handle":"chrisalbon.com","displayName":"Chris Albon","avatar":"https://cdn.bsky.app/img/avatar/plain/did:plc:umpsiyampiq3bpgce7kigydz/bafkreiadswkm4njoz5fsp3irzxmhh6h2z6jmvpblh72zbp2uu7z2iexp3e@jpeg","associated":{"chat":{"allowIncoming":"all"}},"viewer":{"muted":false,"blockedBy":false,"following":"at://did:plc:5lwweotr4gfb7bbz2fqwdthf/app.bsky.graph.follow/3l7r3fg3shx2g","followedBy":"at://did:plc:umpsiyampiq3bpgce7kigydz/app.bsky.graph.follow/3l7l72e7da22a"},"labels":[],"createdAt":"2023-03-16T01:40:20.177Z"},"joinedAllTimeCount":0,"joinedWeekCount":0,"labels":[],"indexedAt":"2024-10-31T19:11:16.971Z"},{"uri":"at://did:plc:umpsiyampiq3bpgce7kigydz/app.bsky.graph.starterpack/3l7qpbmatnb2c","cid":"bafyreieyngelz5dwe3x5ktlbegqqaa7n5ni53j4xjbcltseko324gfsqqu","record":{"$type":"app.bsky.graph.starterpack","createdAt":"2024-10-30T17:48:27.214Z","list":"at://did:plc:umpsiyampiq3bpgce7kigydz/app.bsky.graph.list/3l7qpblunwu22","name":"Chris Albon's Starter Pack"},"creator":{"did":"did:plc:umpsiyampiq3bpgce7kigydz","handle":"chrisalbon.com","displayName":"Chris Albon","avatar":"https://cdn.bsky.app/img/avatar/plain/did:plc:umpsiyampiq3bpgce7kigydz/bafkreiadswkm4njoz5fsp3irzxmhh6h2z6jmvpblh72zbp2uu7z2iexp3e@jpeg","associated":{"chat":{"allowIncoming":"all"}},"viewer":{"muted":false,"blockedBy":false,"following":"at://did:plc:5lwweotr4gfb7bbz2fqwdthf/app.bsky.graph.follow/3l7r3fg3shx2g","followedBy":"at://did:plc:umpsiyampiq3bpgce7kigydz/app.bsky.graph.follow/3l7l72e7da22a"},"labels":[],"createdAt":"2023-03-16T01:40:20.177Z"},"joinedAllTimeCount":0,"joinedWeekCount":0,"labels":[],"indexedAt":"2024-10-30T17:48:27.214Z"}]}`
	jsonString := `{"starterPacks":[{"uri":"at://did:plc:p7uix7mresfq4nfzxp3klgfa/app.bsky.graph.starterpack/3l74pjn6n3d26","cid":"bafyreia4gz3otb4zeoyyubmnpculrq2l6lnpiod5wtt6jsnqb3kaxwv4qq","record":{"$type":"app.bsky.graph.starterpack","createdAt":"2024-10-22T18:59:41.873Z","description":"People speaking at https://2024.allthingsopen.org","descriptionFacets":[{"features":[{"$type":"app.bsky.richtext.facet#link","uri":"https://2024.allthingsopen.org"}],"index":{"byteEnd":49,"byteStart":19}}],"feeds":[{"cid":"bafyreiac75pryims2jtd375t7vqluhej3a7d67ps3wqu5eleloe665id2q","creator":{"associated":{"chat":{"allowIncoming":"none"}},"avatar":"https://cdn.bsky.app/img/avatar/plain/did:plc:p7uix7mresfq4nfzxp3klgfa/bafkreicvbh4pdflctephx3ptiruxfr23gl4voguklcoekmt6q37vdkwvdi@jpeg","createdAt":"2023-04-22T20:11:47.863Z","description":"Futurist historian\n\nDM Signal @justin.13","did":"did:plc:p7uix7mresfq4nfzxp3klgfa","displayName":"Justin Garrison","handle":"justingarrison.com","indexedAt":"2024-10-22T22:29:51.542Z","labels":[],"viewer":{"blockedBy":false,"muted":false}},"description":"All Things Open","did":"did:web:skyfeed.me","displayName":"All Things Open","indexedAt":"2024-10-22T20:34:28.736Z","labels":[],"likeCount":2,"uri":"at://did:plc:p7uix7mresfq4nfzxp3klgfa/app.bsky.feed.generator/aaagsmpdw3lza","viewer":{}}],"list":"at://did:plc:p7uix7mresfq4nfzxp3klgfa/app.bsky.graph.list/3l74pjmruut26","name":"All Things Open 2024 speakers","updatedAt":"2024-10-23T16:22:11.025Z"},"creator":{"did":"did:plc:p7uix7mresfq4nfzxp3klgfa","handle":"justingarrison.com","displayName":"Justin Garrison","avatar":"https://cdn.bsky.app/img/avatar/plain/did:plc:p7uix7mresfq4nfzxp3klgfa/bafkreicvbh4pdflctephx3ptiruxfr23gl4voguklcoekmt6q37vdkwvdi@jpeg","associated":{"chat":{"allowIncoming":"none"}},"viewer":{"muted":false,"blockedBy":false,"following":"at://did:plc:5lwweotr4gfb7bbz2fqwdthf/app.bsky.graph.follow/3l7tjyh4nft2b"},"labels":[],"createdAt":"2023-04-22T20:11:47.863Z"},"joinedAllTimeCount":0,"joinedWeekCount":0,"labels":[],"indexedAt":"2024-10-22T18:59:41.873Z"},{"uri":"at://did:plc:p7uix7mresfq4nfzxp3klgfa/app.bsky.graph.starterpack/3kvwk4rncwb2k","cid":"bafyreid7tfhi4xbyttotunwj2xxbxvpsznyhsa33ow2fefvvvb2x76ykxq","record":{"$type":"app.bsky.graph.starterpack","createdAt":"2024-06-27T19:20:18.501Z","description":"People involved in Cloud Native and Kubernetes ecosystem","feeds":[],"list":"at://did:plc:p7uix7mresfq4nfzxp3klgfa/app.bsky.graph.list/3kvwk4rg5fe2j","name":"Cloud Native","updatedAt":"2024-11-03T00:41:26.923Z"},"creator":{"did":"did:plc:p7uix7mresfq4nfzxp3klgfa","handle":"justingarrison.com","displayName":"Justin Garrison","avatar":"https://cdn.bsky.app/img/avatar/plain/did:plc:p7uix7mresfq4nfzxp3klgfa/bafkreicvbh4pdflctephx3ptiruxfr23gl4voguklcoekmt6q37vdkwvdi@jpeg","associated":{"chat":{"allowIncoming":"none"}},"viewer":{"muted":false,"blockedBy":false,"following":"at://did:plc:5lwweotr4gfb7bbz2fqwdthf/app.bsky.graph.follow/3l7tjyh4nft2b"},"labels":[],"createdAt":"2023-04-22T20:11:47.863Z"},"joinedAllTimeCount":19,"joinedWeekCount":1,"labels":[],"indexedAt":"2024-06-27T19:20:18.501Z"}]}`

	unstructured := map[string]interface{}{}
	if err := json.Unmarshal([]byte(jsonString), &unstructured); err != nil {
		t.Fatalf("json.Unmarshal() = %v, wanted nil", err)
	}
	fmt.Printf("jsonString:\n%s\n", helpers.PrettyString(unstructured))
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

func Test_SyncStarterPackToFile(t *testing.T) {
	if os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skipf("Test_StarterPacks is a manual test that is skipped in CICD")
	}

	stuff, err := testSetup()
	if err != nil {
		t.Fatalf("testSetup() = %v, wanted nil", err)
	}

	// https://bsky.app/starter-pack-short/RCerxDE
	//handle := "justingarrison.com"
	// startPackName := "Cloud Native"
	//handle := "bryanross.me"
	//startPackName := "Platform Engineering Starter Pack"
	//sourceFile := "/Users/jlewi/git_bskylists/platformengineering.yaml"

	sourceFile := "/Users/jlewi/git_bskylists/aiengineering.yaml"
	handle := "nimobeeren.com"
	startPackName := "AI/ML/data frens starter pack"

	err = func() error {
		b, err := os.ReadFile(sourceFile)
		if err != nil {
			return errors.Wrapf(err, "cannot read file %s", sourceFile)
		}

		nodes, err := kio.FromBytes(b)
		if err != nil {
			return errors.Wrapf(err, "cannot read file %s", sourceFile)
		}

		node := nodes[0]
		dest := &FollowList{}
		if err := node.YNode().Decode(dest); err != nil {
			return errors.Wrapf(err, "cannot unmarshal FollowList from file %s", sourceFile)
		}

		output, err := DumpStarterPack(stuff.Client, handle, startPackName)
		if err != nil {
			return err
		}

		MergeFollowLists(dest, *output)

		outB, err := yaml.Marshal(dest)
		if err != nil {
			return errors.Wrapf(err, "cannot marshal FollowList to file %s", sourceFile)
		}

		if err := os.WriteFile(sourceFile, outB, 0644); err != nil {
			return errors.Wrapf(err, "cannot write file %s", sourceFile)
		}
		return nil
	}()
	if err != nil {
		t.Fatalf("SyncStarterPackToFile Failed; error %+v", err)
	}
}
