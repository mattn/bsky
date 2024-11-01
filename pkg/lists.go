package pkg

import (
	"context"
	"fmt"
	comatproto "github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	lexutil "github.com/bluesky-social/indigo/lex/util"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/go-logr/zapr"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"time"
)

const (
	referenceListPurpose = "app.bsky.graph.defs#referencelist"
)

// ListRecord struct to represent the list record structure
type ListRecord struct {
	Type        string    `json:"$type"`
	CreatedAt   time.Time `json:"createdAt"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Users       []User    `json:"users"`
}

// User struct to represent each user in the list
type User struct {
	DID string `json:"did"`
}

//
//func GetList(client *xrpc.Client, listAtRef string) {
//	//TODO(jeremy): Support the cursor?
//	cursor := ""
//	limit := int64(100)
//	bsky.GraphGetList(context.Background(), client, cursor, limit, listAtRef)
//}

// CreateListRecord sends a request to the PDS server to create a list record.
func CreateListRecord(client *xrpc.Client, record *ListRecord) error {
	//var out bytes.Buffer
	//if err := client.Do(context.Background(), xrpc.Procedure, "", "app.bsky.graph.list", nil, record, &out); err != nil {
	//	return err
	//}

	description := "Test programmatically creating a list"
	// I think we need to create a list and then we create GraphListItem
	block := bsky.GraphList{
		LexiconTypeID: "app.bsky.graph.list",
		CreatedAt:     time.Now().Local().Format(time.RFC3339),
		Name:          "TestList",
		Description:   &description,
		Purpose:       StringPtr(referenceListPurpose),
	}

	//block := bsky.GraphBlock{
	//	LexiconTypeID: "app.bsky.graph.block",
	//	CreatedAt:     time.Now().Local().Format(time.RFC3339),
	//	Subject:       profile.Did,
	//}

	resp, err := comatproto.RepoCreateRecord(context.TODO(), client, &comatproto.RepoCreateRecord_Input{
		Collection: "app.bsky.graph.list",
		Repo:       client.Auth.Did,
		Record: &lexutil.LexiconTypeDecoder{
			Val: &block,
		},
	})

	fmt.Printf("List record created:\n%v", resp)

	if err != nil {
		return err
	}

	item := bsky.GraphListitem{
		LexiconTypeID: "app.bsky.graph.listitem",
		CreatedAt:     time.Now().Local().Format(time.RFC3339),
		List:          resp.Uri,
		Subject:       "did:plc:umpsiyampiq3bpgce7kigydz",
	}

	//block := bsky.GraphBlock{
	//	LexiconTypeID: "app.bsky.graph.block",
	//	CreatedAt:     time.Now().Local().Format(time.RFC3339),
	//	Subject:       profile.Did,
	//}

	itemResp, err := comatproto.RepoCreateRecord(context.TODO(), client, &comatproto.RepoCreateRecord_Input{
		Collection: "app.bsky.graph.listitem",
		Repo:       client.Auth.Did,
		Record: &lexutil.LexiconTypeDecoder{
			Val: &item,
		},
	})

	if err != nil {
		return err
	}
	fmt.Printf("List item record created:\n%v", itemResp)
	return nil
}

// AddToList adds a subjectDid to the list
func AddToList(client *xrpc.Client, listURI string, subjectDid string) error {
	item := bsky.GraphListitem{
		LexiconTypeID: "app.bsky.graph.listitem",
		CreatedAt:     time.Now().Local().Format(time.RFC3339),
		List:          listURI,
		Subject:       subjectDid,
	}

	//block := bsky.GraphBlock{
	//	LexiconTypeID: "app.bsky.graph.block",
	//	CreatedAt:     time.Now().Local().Format(time.RFC3339),
	//	Subject:       profile.Did,
	//}

	itemResp, err := comatproto.RepoCreateRecord(context.TODO(), client, &comatproto.RepoCreateRecord_Input{
		Collection: "app.bsky.graph.listitem",
		Repo:       client.Auth.Did,
		Record: &lexutil.LexiconTypeDecoder{
			Val: &item,
		},
	})
	log := zapr.NewLogger(zap.L())
	log.Info("List item record created", "item", itemResp)

	return errors.Wrapf(err, "Failed to add subject to list: %s", listURI)
}

//func MutateList(client *xrpc.Client, listRef string) error {
//	cursor := ""
//	limit := int64(100)
//	list, err := bsky.GraphGetList(context.Background(), client, cursor, limit, listRef)
//
//	if err != nil {
//		return errors.Wrapf(err, "Failed to fetch list: %s", listRef)
//	}
//
//	itemResp, err := comatproto.RepoPutRecord(context.TODO(), client, &comatproto.RepoCreateRecord_Input{
//		Collection: "app.bsky.graph.listitem",
//		Repo:       client.Auth.Did,
//		Record: &lexutil.LexiconTypeDecoder{
//			Val: &item,
//		},
//	})
//}
