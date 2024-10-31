package pkg

import (
	"os"
	"testing"
)

func Test_StarterPacks(t *testing.T) {
	if os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skipf("Test_StarterPacks is a manual test that is skipped in CICD")
	}

}
