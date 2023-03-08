package main

import (
	"testing"
)

func TestTimep(t *testing.T) {
	want := "2023-02-03T18:19:20.333Z"
	got := timep(want).UTC().Format("2006-01-02T15:04:05.000Z")
	if got != want {
		t.Fatalf("want %q but got %q", want, got)
	}

	want = "2023-02-03T18:19:20"
	got = timep(want).UTC().Format("2006-01-02T15:04:05.000Z")
	if got == want {
		t.Fatal("should not be possible to parse")
	}
}

func TestStringp(t *testing.T) {
	want := "test"
	got := stringp(&want)
	if got != want {
		t.Fatalf("want %q but got %q", want, got)
	}

	want = ""
	got = stringp(nil)
	if got != want {
		t.Fatalf("want %q but got %q", want, got)
	}
}
