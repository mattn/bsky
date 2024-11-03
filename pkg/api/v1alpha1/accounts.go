package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	AccountListKind = "AccountList"
	AccountListGVK  = schema.FromAPIVersionAndKind(Group+"/"+Version, AccountListKind)
)

// AccountList is a data structure to hold a list of folks to follow
type AccountList struct {
	APIVersion string   `json:"apiVersion" yaml:"apiVersion"`
	Kind       string   `json:"kind" yaml:"kind"`
	Metadata   Metadata `json:"metadata" yaml:"metadata"`

	// DID is the Decentralized Identifier for the list
	// TOOD(jeremy):
	DID      string    `json:"did" yaml:"did"`
	Accounts []Account `json:"accounts" yaml:"accounts"`
}

type Account struct {
	Handle string `json:"handle" yaml:"handle"`
}
