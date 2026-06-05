package main

import "testing"

func TestStreamHost(t *testing.T) {
	tests := []struct {
		name   string
		host   string
		cursor string
		want   string
	}{
		{
			name: "without cursor",
			host: "https://bsky.network",
			want: "wss://bsky.network/xrpc/com.atproto.sync.subscribeRepos",
		},
		{
			name:   "with cursor",
			host:   "https://bsky.network",
			cursor: "123",
			want:   "wss://bsky.network/xrpc/com.atproto.sync.subscribeRepos?cursor=123",
		},
		{
			name:   "preserve existing query",
			host:   "https://bsky.network?foo=bar",
			cursor: "123",
			want:   "wss://bsky.network/xrpc/com.atproto.sync.subscribeRepos?cursor=123&foo=bar",
		},
	}

	for _, tt := range tests {
		got, err := streamHost(tt.host, tt.cursor)
		if err != nil {
			t.Fatalf("%s: streamHost returned error: %v", tt.name, err)
		}
		if got != tt.want {
			t.Fatalf("%s: want %q but got %q", tt.name, tt.want, got)
		}
	}
}
