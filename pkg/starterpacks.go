package pkg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/bluesky-social/indigo/xrpc"
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
}

// Record struct representing the inner "record" field in each starter pack
type Record struct {
	Type      string    `json:"$type"`
	CreatedAt time.Time `json:"createdAt"`
	Feeds     []string  `json:"feeds"`
	List      string    `json:"list"`
	Name      string    `json:"name"`
}

// Creator struct representing the "creator" field in each starter pack
type Creator struct {
	DID         string     `json:"did"`
	Handle      string     `json:"handle"`
	DisplayName string     `json:"displayName"`
	Avatar      string     `json:"avatar"`
	Associated  Associated `json:"associated"`
	Viewer      Viewer     `json:"viewer"`
	Labels      []string   `json:"labels"`
	CreatedAt   time.Time  `json:"createdAt"`
}

// Associated struct within "creator"
type Associated struct {
	Chat ChatSettings `json:"chat"`
}

// ChatSettings struct within "associated" to specify chat permissions
type ChatSettings struct {
	AllowIncoming string `json:"allowIncoming"`
}

// Viewer struct representing viewer's interactions with the creator
type Viewer struct {
	Muted      bool   `json:"muted"`
	BlockedBy  bool   `json:"blockedBy"`
	Following  string `json:"following"`
	FollowedBy string `json:"followedBy"`
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
	var out GetStarterPacks_Output
	if err := client.Do(context.Background(), xrpc.Query, "", "app.bsky.graph.getActorStarterPacks", params, nil, &out); err != nil {
		return out, err
	}

	return out, nil
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
