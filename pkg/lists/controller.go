package lists

import (
	"context"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/jlewi/bsctl/pkg/api/v1alpha1"
	"github.com/pkg/errors"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type AccountListController struct {
	client *xrpc.Client
}

func NewAccountListController(client *xrpc.Client) (*AccountListController, error) {
	return &AccountListController{
		client: client,
	}, nil
}

func (c *AccountListController) ReconcileNode(ctx context.Context, n *yaml.RNode) error {
	list := &v1alpha1.AccountList{}
	if err := n.YNode().Decode(list); err != nil {
		return errors.Wrapf(err, "Failed to decode AccountList")
	}

	return c.Reconcile(ctx, list)
}

func (c *AccountListController) Reconcile(ctx context.Context, list *v1alpha1.AccountList) error {
	if list.DID == "" {
		return errors.New("List did must be specified. We currently don't support creating new lists yet.")
	}
	if err := AddAllToList(c.client, list.DID, *list); err != nil {
		return err
	}
	return nil
}
