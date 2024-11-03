package v1alpha1

import "k8s.io/apimachinery/pkg/runtime/schema"

var (
	FeedKind = "Feed"
	FeedGVK  = schema.FromAPIVersionAndKind(Group+"/"+Version, FeedKind)
)

// Feed is a data structure to hold the configuration for a feed using blueskyfeedcreator
type Feed struct {
	APIVersion string   `json:"apiVersion" yaml:"apiVersion"`
	Kind       string   `json:"kind" yaml:"kind"`
	Metadata   Metadata `json:"metadata" yaml:"metadata"`
	// Include is a list of the terms to match for posts to be included.
	Include []string `json:"include" yaml:"include"`
}
