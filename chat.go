package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/api/chat"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

type didDocument struct {
	Service []didService `json:"service"`
}

type didService struct {
	ID              string `json:"id"`
	Type            string `json:"type"`
	ServiceEndpoint string `json:"serviceEndpoint"`
}

func resolvePDS(did string) (string, error) {
	var url string
	if strings.HasPrefix(did, "did:plc:") {
		url = "https://plc.directory/" + did
	} else if strings.HasPrefix(did, "did:web:") {
		host := strings.TrimPrefix(did, "did:web:")
		url = "https://" + host + "/.well-known/did.json"
	} else {
		return "", fmt.Errorf("unsupported DID method: %s", did)
	}

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("cannot resolve DID: %w", err)
	}
	defer resp.Body.Close()

	var doc didDocument
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return "", fmt.Errorf("cannot decode DID document: %w", err)
	}

	for _, svc := range doc.Service {
		if svc.ID == "#atproto_pds" || svc.Type == "AtprotoPersonalDataServer" {
			return svc.ServiceEndpoint, nil
		}
	}
	return "", fmt.Errorf("PDS service not found in DID document")
}

func makeChatXRPCC(cCtx *cli.Context) (*xrpc.Client, error) {
	xrpcc, err := makeXRPCC(cCtx)
	if err != nil {
		return nil, err
	}

	// Chat API requires proxying through the actual PDS, not the entryway
	pdsHost, err := resolvePDS(xrpcc.Auth.Did)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve PDS: %w", err)
	}
	xrpcc.Host = pdsHost
	xrpcc.Headers = map[string]string{
		"Atproto-Proxy": "did:web:api.bsky.chat#bsky_chat",
	}
	return xrpcc, nil
}

func doConvos(cCtx *cli.Context) error {
	if cCtx.Args().Present() {
		return cli.ShowSubcommandHelp(cCtx)
	}

	xrpcc, err := makeChatXRPCC(cCtx)
	if err != nil {
		return fmt.Errorf("cannot create client: %w", err)
	}

	var cursor string
	for {
		resp, err := chat.ConvoListConvos(context.TODO(), xrpcc, cursor, 50, "", "")
		if err != nil {
			return fmt.Errorf("cannot list conversations: %w", err)
		}

		if cCtx.Bool("json") {
			for _, c := range resp.Convos {
				json.NewEncoder(os.Stdout).Encode(c)
			}
		} else {
			for _, c := range resp.Convos {
				var members []string
				for _, m := range c.Members {
					members = append(members, m.Handle)
				}
				color.Set(color.FgHiRed)
				fmt.Print(strings.Join(members, ", "))
				color.Set(color.Reset)
				fmt.Printf(" (unread: %d)", c.UnreadCount)
				if c.LastMessage != nil && c.LastMessage.ConvoDefs_MessageView != nil {
					msg := c.LastMessage.ConvoDefs_MessageView
					text := msg.Text
					if len(text) > 50 {
						text = text[:50] + "..."
					}
					fmt.Printf(" %s", text)
				}
				fmt.Println()
				fmt.Print(" - ")
				color.Set(color.FgBlue)
				fmt.Println(c.Id)
				color.Set(color.Reset)
			}
		}

		if resp.Cursor == nil {
			break
		}
		cursor = *resp.Cursor
	}
	return nil
}

func doConvo(cCtx *cli.Context) error {
	if !cCtx.Args().Present() {
		return cli.ShowSubcommandHelp(cCtx)
	}

	xrpcc, err := makeChatXRPCC(cCtx)
	if err != nil {
		return fmt.Errorf("cannot create client: %w", err)
	}

	convoId := cCtx.Args().First()
	n := cCtx.Int64("n")

	resp, err := chat.ConvoGetMessages(context.TODO(), xrpcc, convoId, "", n)
	if err != nil {
		return fmt.Errorf("cannot get messages: %w", err)
	}

	if cCtx.Bool("json") {
		for _, m := range resp.Messages {
			json.NewEncoder(os.Stdout).Encode(m)
		}
		return nil
	}

	// reverse to show oldest first
	for i, j := 0, len(resp.Messages)-1; i < j; i, j = i+1, j-1 {
		resp.Messages[i], resp.Messages[j] = resp.Messages[j], resp.Messages[i]
	}

	for _, m := range resp.Messages {
		if m.ConvoDefs_MessageView != nil {
			msg := m.ConvoDefs_MessageView
			color.Set(color.FgHiRed)
			fmt.Print(msg.Sender.Did)
			color.Set(color.Reset)
			fmt.Printf(" (%s)\n", msg.SentAt)
			fmt.Println(msg.Text)
			fmt.Println()
		} else if m.ConvoDefs_DeletedMessageView != nil {
			fmt.Println("[deleted message]")
			fmt.Println()
		}
	}
	return nil
}

func doChat(cCtx *cli.Context) error {
	if cCtx.Args().Len() < 2 {
		return cli.ShowSubcommandHelp(cCtx)
	}

	handle := cCtx.Args().First()
	text := strings.Join(cCtx.Args().Slice()[1:], " ")

	// resolve handle to DID using the normal client (not chat-proxied)
	var did string
	if strings.HasPrefix(handle, "did:") {
		did = handle
	} else {
		xrpcc, err := makeXRPCC(cCtx)
		if err != nil {
			return fmt.Errorf("cannot create client: %w", err)
		}
		profile, err := bsky.ActorGetProfile(context.TODO(), xrpcc, handle)
		if err != nil {
			return fmt.Errorf("cannot get profile: %w", err)
		}
		did = profile.Did
	}

	xrpcc, err := makeChatXRPCC(cCtx)
	if err != nil {
		return fmt.Errorf("cannot create client: %w", err)
	}

	// get or create conversation
	convoResp, err := chat.ConvoGetConvoForMembers(context.TODO(), xrpcc, []string{did})
	if err != nil {
		return fmt.Errorf("cannot get conversation: %w", err)
	}

	// send message
	msgResp, err := chat.ConvoSendMessage(context.TODO(), xrpcc, &chat.ConvoSendMessage_Input{
		ConvoId: convoResp.Convo.Id,
		Message: &chat.ConvoDefs_MessageInput{
			Text: text,
		},
	})
	if err != nil {
		return fmt.Errorf("cannot send message: %w", err)
	}

	fmt.Printf("Message sent (id: %s)\n", msgResp.Id)
	return nil
}
