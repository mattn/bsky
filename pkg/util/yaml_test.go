package util

import (
	"os"
	"path"
	"path/filepath"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type entry struct {
	source string
	dest   string
}

func new(s string, d string) entry {
	return entry{
		source: s,
		dest:   d,
	}
}

func Test_FindYamlFiles(t *testing.T) {
	entries := []entry{
		new("", "some/path/notalink.yaml"),
		new("some/path/source.yaml", "some/otherpath/link.yaml"),
	}

	baseDir, err := os.MkdirTemp("", "FindYamlFiles")
	t.Logf("Base Dir: %v", baseDir)
	if err != nil {
		t.Fatalf("Failed to create tempdir; %v", err)
	}

	for _, e := range entries {
		fToCreate := e.source
		if e.source == "" {
			fToCreate = e.dest
		}

		dir := path.Dir(fToCreate)
		tDir := filepath.Join(baseDir, dir)
		t.Logf("Dir: %v", tDir)
		if err := os.MkdirAll(tDir, FilePermUserGroup); err != nil {
			t.Fatalf("Failed to make directory: %v; error:%v", tDir, err)
		}

		fullPath := filepath.Join(baseDir, fToCreate)
		f, err := os.Create(fullPath)
		if err != nil {
			t.Fatalf("Failed to create file: %v; error:%v", fullPath, err)
		}
		if err := f.Close(); err != nil {
			t.Fatalf("Failed to close file: %v; error:%v", fToCreate, err)
		}

		if e.source != "" {
			dir := path.Dir(e.dest)
			tDir := filepath.Join(baseDir, dir)
			t.Logf("Dir: %v", tDir)
			if err := os.MkdirAll(tDir, FilePermUserGroup); err != nil {
				t.Fatalf("Failed to make directory: %v; error:%v", tDir, err)
			}

			fullSource := filepath.Join(baseDir, e.source)
			fullDest := filepath.Join(baseDir, e.dest)
			if err := os.Symlink(fullSource, fullDest); err != nil {
				t.Fatalf("Failed to create symbolic link; source: %v dest: %v; error%v", fullSource, fullDest, err)
			}
		}
	}

	type testCase struct {
		expected []string
	}

	cases := []testCase{
		{
			expected: []string{"some/path/notalink.yaml", "some/path/source.yaml"},
		},
	}

	for _, c := range cases {
		// results, err := FindYamlFiles(baseDir, c.ignore)
		results, err := FindYamlFiles(baseDir)
		if err != nil {
			t.Errorf("FindYamlFiles returned error: %v", err)
		}
		fullExpected := []string{}
		for _, e := range c.expected {
			eResolved, err := filepath.EvalSymlinks(filepath.Join(baseDir, e))
			if err != nil {
				t.Fatalf("Could not evaluate symlink %v; err %v", e, err)
			}
			fullExpected = append(fullExpected, eResolved)
		}

		sort.Strings(fullExpected)
		sort.Strings(results)

		d := cmp.Diff(fullExpected, results)

		if d != "" {
			t.Errorf("FindYamlFiles didn't return expected:\n%v", d)
		}

	}
}
