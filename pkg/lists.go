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
	"sort"
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
//
// TODO(jeremy): How should we check if a list of the given name already exists?
func CreateListRecord(client *xrpc.Client, name string, description string) (*comatproto.RepoCreateRecord_Output, error) {
	//var out bytes.Buffer
	//if err := client.Do(context.Background(), xrpc.Procedure, "", "app.bsky.graph.list", nil, record, &out); err != nil {
	//	return err
	//}
	log := zapr.NewLogger(zap.L())

	// I think we need to create a list and then we create GraphListItem
	block := bsky.GraphList{
		LexiconTypeID: "app.bsky.graph.list",
		CreatedAt:     time.Now().Local().Format(time.RFC3339),
		Name:          name,
		Description:   StringPtr(description),
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

	if err != nil {
		log.Error(err, "Failed to create list record")
		return nil, err
	}

	log.Info("List record created", "record", resp)

	return resp, nil
}

//func GetList(client *xrpc.Client, listUri string) (*comatproto.RepoCreateRecord_Output, error) {
//	//var out bytes.Buffer
//	//if err := client.Do(context.Background(), xrpc.Procedure, "", "app.bsky.graph.list", nil, record, &out); err != nil {
//	//	return err
//	//}
//	log := zapr.NewLogger(zap.L())
//
//	//// I think we need to create a list and then we create GraphListItem
//	//block := bsky.GraphList{
//	//	LexiconTypeID: "app.bsky.graph.list",
//	//	CreatedAt:     time.Now().Local().Format(time.RFC3339),
//	//	Name:          name,
//	//	Description:   StringPtr(description),
//	//	Purpose:       StringPtr(referenceListPurpose),
//	//}
//
//	//block := bsky.GraphBlock{
//	//	LexiconTypeID: "app.bsky.graph.block",
//	//	CreatedAt:     time.Now().Local().Format(time.RFC3339),
//	//	Subject:       profile.Did,
//	//}
//
//	bsky.GraphGetList()
//	resp, err := comatproto.G(context.TODO(), client, &comatproto.RepoCreateRecord_Input{
//		Collection: "app.bsky.graph.list",
//		Repo:       client.Auth.Did,
//		Record: &lexutil.LexiconTypeDecoder{
//			Val: &block,
//		},
//	})
//
//	if err != nil {
//		log.Error(err, "Failed to create list record")
//		return nil, err
//	}
//
//	log.Info("List record created", "record", resp)
//
//	return resp, nil
//}

func AddAllToList(client *xrpc.Client, listURI string, source FollowList) error {
	log := zapr.NewLogger(zap.L())
	for _, h := range source.Accounts {
		profile, err := bsky.ActorGetProfile(context.TODO(), client, h.Handle)
		if err != nil {
			var xErr *xrpc.Error
			if errors.As(err, &xErr) {
				if 400 == xErr.StatusCode {
					log.Error(err, "Profile not found for handle", "handle", h)
					continue
				}
			}
			return fmt.Errorf("cannot get profile: %w", err)
		}
		AddToList(client, listURI, profile.Did)
	}

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

// MergeFollowLists computes the union of two lists
func MergeFollowLists(dest *FollowList, src FollowList) {
	// Use a map to store unique strings from both lists
	uniqueStrings := make(map[string]bool)

	// Add elements from the first list to the map
	for _, item := range dest.Accounts {
		uniqueStrings[item.Handle] = true
	}

	// Add elements from the second list to the map
	for _, item := range src.Accounts {
		uniqueStrings[item.Handle] = true
	}

	// Convert map keys to a slice
	result := make([]string, 0, len(uniqueStrings))
	for item := range uniqueStrings {
		result = append(result, item)
	}

	// Sort the result slice
	sort.Strings(result)

	dest.Accounts = make([]Account, 0, len(result))
	for _, item := range result {
		dest.Accounts = append(dest.Accounts, Account{Handle: item})
	}

}
