package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// StarterPackRecord represents the structure of a starter pack record
type StarterPackRecord struct {
	Profiles []string `json:"profiles"` // list of DID (Decentralized Identifier) strings
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
