package lists

import (
	"context"
	"github.com/jlewi/bsctl/pkg/api/v1alpha1"
	"github.com/pkg/errors"
	"sigs.k8s.io/kustomize/kyaml/yaml"
	"sort"
)

type FeedController struct {
}

func NewFeedController() (*FeedController, error) {
	return &FeedController{}, nil
}

func (c *FeedController) ReconcileNode(_ context.Context, _ *yaml.RNode) error {
	// Its a null op for now
	return nil
}

func (c *FeedController) TidyNode(ctx context.Context, n *yaml.RNode) (*yaml.RNode, error) {
	f := &v1alpha1.Feed{}
	if err := n.YNode().Decode(f); err != nil {
		return nil, errors.Wrapf(err, "Failed to decode Feed")
	}

	if err := c.Tidy(ctx, f); err != nil {
		return nil, err
	}

	if err := n.YNode().Encode(f); err != nil {
		return nil, errors.Wrapf(err, "Failed to encode Feed")
	}
	return n, nil
}

func (c *FeedController) Tidy(ctx context.Context, f *v1alpha1.Feed) error {
	// dedupe include terms
	terms := make(map[string]bool)
	for _, term := range f.Include {
		terms[term] = true
	}
	final := make([]string, 0, len(terms))
	for term := range terms {
		final = append(final, term)
	}
	sort.Strings(final)
	f.Include = final
	return nil
}
