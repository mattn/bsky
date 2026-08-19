package main

import (
	"bytes"
	"image"
	"image/png"
	"math/rand"
	"testing"
	"time"
)

func TestTimep(t *testing.T) {
	want := "2023-02-03T18:19:20Z"
	got := timep(want).UTC().Format(time.RFC3339)
	if got != want {
		t.Fatalf("want %q but got %q", want, got)
	}

	want = "2023-02-03T18:19:20.333Z"
	got = timep(want).UTC().Format(time.RFC3339)
	if got == want {
		t.Fatalf("want %q but got %q", want, got)
	}

	want = "2023-02-03T18:19:20"
	got = timep(want).UTC().Format(time.RFC3339)
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

func TestCompressImage(t *testing.T) {
	// A small image is returned unchanged.
	var buf bytes.Buffer
	small := image.NewRGBA(image.Rect(0, 0, 10, 10))
	if err := png.Encode(&buf, small); err != nil {
		t.Fatal(err)
	}
	b, mimeType, err := compressImage(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(b, buf.Bytes()) {
		t.Fatal("small image should be returned unchanged")
	}
	if mimeType != "image/png" {
		t.Fatalf("want %q but got %q", "image/png", mimeType)
	}

	// A large noisy image is re-encoded to fit in maxBlobSize.
	large := image.NewRGBA(image.Rect(0, 0, 2000, 2000))
	r := rand.New(rand.NewSource(1))
	for i := range large.Pix {
		large.Pix[i] = uint8(r.Intn(256))
	}
	buf.Reset()
	if err := png.Encode(&buf, large); err != nil {
		t.Fatal(err)
	}
	if buf.Len() <= maxBlobSize {
		t.Fatalf("test image should be larger than %d but got %d", maxBlobSize, buf.Len())
	}
	b, mimeType, err = compressImage(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if len(b) > maxBlobSize {
		t.Fatalf("compressed image should be smaller than %d but got %d", maxBlobSize, len(b))
	}
	if mimeType != "image/jpeg" {
		t.Fatalf("want %q but got %q", "image/jpeg", mimeType)
	}

	// Broken data is an error.
	if _, _, err = compressImage(bytes.Repeat([]byte{0}, maxBlobSize+1)); err == nil {
		t.Fatal("broken image should be an error")
	}
}
