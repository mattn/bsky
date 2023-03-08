package main

import (
	"reflect"
	"testing"
)

func TestExtractLinks(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []entry
	}{
		{name: "1", input: `検索は https://google.com です`, want: []entry{{text: "https://google.com", start: 4, end: 22}}},
		{name: "2", input: `https://google.com です`, want: []entry{{text: "https://google.com", start: 0, end: 18}}},
		{name: "3", input: `https://google.com`, want: []entry{{text: "https://google.com", start: 0, end: 18}}},
	}
	for _, test := range tests {
		result := extractLinks(test.input)
		if len(result) != len(test.want) {
			t.Fatalf("extract %d link(s)", len(test.want))
		}
		if !reflect.DeepEqual(result, test.want) {
			t.Fatalf("want %v but got %v for test %v", test.want, result, test.name)
		}
	}
}

func TestExtractMentions(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []entry
	}{
		{name: "1", input: `返事は @mattn へ`, want: []entry{{text: "mattn", start: 4, end: 10}}},
		{name: "2", input: `返事は @mattn-- へ`, want: []entry{{text: "mattn", start: 4, end: 10}}},
		{name: "3", input: `返事は @mattn.jp へ`, want: []entry{{text: "mattn.jp", start: 4, end: 13}}},
		{name: "4", input: `返事は @@mattn へ`, want: []entry{{text: "mattn", start: 5, end: 11}}},
	}
	for _, test := range tests {
		result := extractMentions(test.input)
		if len(result) != len(test.want) {
			t.Fatalf("extract %d link(s)", len(test.want))
		}
		if !reflect.DeepEqual(result, test.want) {
			t.Fatalf("want %v but got %v for test %v", test.want, result, test.name)
		}
	}
}
