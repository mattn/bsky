package pkg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/go-logr/zapr"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net/http"
	"time"
)

// StarterPackRecord represents the structure of a starter pack record
type StarterPackRecord struct {
	Profiles []string `json:"profiles"` // list of DID (Decentralized Identifier) strings
}

// GetStarterPacks_Output is the output of a app.bsky.graph.getFollows call.
type GetStarterPacks_Output struct {
	Cursor       *string        `json:"cursor,omitempty" cborgen:"cursor,omitempty"`
	StarterPacks []*StarterPack `json:"starterPacks" cborgen:"starterPacks"`
}

// Struct representing each starter pack in the "starterPacks" array
type StarterPack struct {
	URI                string    `json:"uri"`
	CID                string    `json:"cid"`
	Record             Record    `json:"record"`
	Creator            Creator   `json:"creator"`
	JoinedAllTimeCount int       `json:"joinedAllTimeCount"`
	JoinedWeekCount    int       `json:"joinedWeekCount"`
	Labels             []string  `json:"labels"`
	IndexedAt          time.Time `json:"indexedAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

// Record struct representing the inner "record" field in each starter pack
type Record struct {
	Type              string             `json:"$type"`
	CreatedAt         time.Time          `json:"createdAt"`
	UpdatedAt         time.Time          `json:"updatedAt"`
	Feeds             []Feed             `json:"feeds"`
	List              string             `json:"list"`
	Name              string             `json:"name"`
	Description       string             `json:"description"`
	DescriptionFacets []DescriptionFacet `json:"descriptionFacets"`
}

type ChatSettings struct {
	AllowIncoming string `json:"allowIncoming"`
}

type Associated struct {
	Chat ChatSettings `json:"chat"`
}

type Viewer struct {
	BlockedBy bool   `json:"blockedBy"`
	Muted     bool   `json:"muted"`
	Following string `json:"following"`
}

type Creator struct {
	Associated  Associated `json:"associated"`
	Avatar      string     `json:"avatar"`
	CreatedAt   time.Time  `json:"createdAt"`
	Description string     `json:"description"`
	Did         string     `json:"did"`
	DisplayName string     `json:"displayName"`
	Handle      string     `json:"handle"`
	IndexedAt   time.Time  `json:"indexedAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	Labels      []string   `json:"labels"`
	Viewer      Viewer     `json:"viewer"`
}

type Feed struct {
	CID         string    `json:"cid"`
	Creator     Creator   `json:"creator"`
	Description string    `json:"description"`
	Did         string    `json:"did"`
	DisplayName string    `json:"displayName"`
	IndexedAt   time.Time `json:"indexedAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Labels      []string  `json:"labels"`
	LikeCount   int       `json:"likeCount"`
	URI         string    `json:"uri"`
	Viewer      struct{}  `json:"viewer"`
}

type Feature struct {
	Type string `json:"$type"`
	URI  string `json:"uri"`
}

type Index struct {
	ByteEnd   int `json:"byteEnd"`
	ByteStart int `json:"byteStart"`
}

type DescriptionFacet struct {
	Features []Feature `json:"features"`
	Index    Index     `json:"index"`
}

// GetStarterPacks retrieves starter packs record from the PDS server.
// https://docs.bsky.app/docs/api/app-bsky-graph-get-actor-starter-packs
func GetStarterPacks(client *xrpc.Client, actor string) (GetStarterPacks_Output, error) {
	limit := 10
	cursor := ""
	params := map[string]interface{}{
		"actor":  actor,
		"cursor": cursor,
		"limit":  limit,
	}
	//var out GetStarterPacks_Output
	var raw bytes.Buffer
	out := GetStarterPacks_Output{}
	if err := client.Do(context.Background(), xrpc.Query, "", "app.bsky.graph.getActorStarterPacks", params, nil, &raw); err != nil {
		return out, err
	}

	if err := json.NewDecoder(&raw).Decode(&out); err != nil {
		log := zapr.NewLogger(zap.L())
		log.Error(err, "Failed to decode JSON response", "response", raw.String())
		return out, err
	}
	return out, nil
}

// DumpStarterPack dumps all the users in a starter pack
func DumpStarterPack(client *xrpc.Client, actor string, name string) (*FollowList, error) {
	out, err := GetStarterPacks(client, actor)
	if err != nil {
		return nil, err
	}

	var starterPack *StarterPack
	for _, pack := range out.StarterPacks {
		if pack.Record.Name == name {
			starterPack = pack
			break
		}
	}

	if starterPack == nil {
		return nil, errors.Errorf("Handle %s doesn't have a starter pack named %s", actor, name)
	}

	// Get the list
	cursor := ""

	result := &FollowList{
		Accounts: make([]Account, 0),
	}
	for {
		output, err := bsky.GraphGetList(context.Background(), client, cursor, 100, starterPack.Record.List)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to get list associated with the starter pack")
		}

		for _, item := range output.Items {
			result.Accounts = append(result.Accounts, Account{Handle: item.Subject.Handle})
		}

		if output.Cursor == nil {
			break
		}
		cursor = *output.Cursor
	}
	return result, nil
}

// CreateStarterPack sends a request to create a starter pack record on the specified PDS server.
func CreateStarterPack(record StarterPackRecord, apiEndpoint, authToken string) error {
	// Marshal the record into JSON
	recordData, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("error marshaling record: %w", err)
	}

	// Create HTTP POST request
	req, err := http.NewRequest("POST", apiEndpoint, bytes.NewBuffer(recordData))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
	req.Header.Set("Content-Type", "application/json")

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to create starter pack record, status: %s", resp.Status)
	}

	return nil
}
