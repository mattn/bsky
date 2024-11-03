package e2etests

import (
	"context"
	"github.com/jlewi/bsctl/pkg/api/v1alpha1"
	"github.com/jlewi/bsctl/pkg/lists"
	"github.com/jlewi/bsctl/pkg/testutil"
	"gopkg.in/yaml.v3"
	"os"
	"testing"
)

func Test_AccountsListApply(t *testing.T) {
	if os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skipf("Test_AccountsListApply is a manual test that is skipped in CICD")
	}

	stuff, err := testutil.New()
	if err != nil {
		t.Fatalf("testSetup() = %v, wanted nil", err)
	}

	sourceFile := "/Users/jlewi/git_bskylists/aiengineering.yaml"

	raw, err := os.ReadFile(sourceFile)
	if err != nil {
		t.Fatalf("Failed to read file; %v; error %+v", sourceFile, err)
	}

	followList := &v1alpha1.AccountList{}
	if err := yaml.Unmarshal(raw, followList); err != nil {
		t.Fatalf("Failed to unmarshal follow list; %v; error %+v", sourceFile, err)
	}

	c, err := lists.NewAccountListController(stuff.Client)
	if err != nil {
		t.Fatalf("NewAccountListController returned error: %+v", err)

	}

	if err := c.Reconcile(context.Background(), followList); err != nil {
		t.Fatalf("Reconcile failed; error: %+v", err)
	}
}
