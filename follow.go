package main

// FollowList is a data structure to hold a list of folks to follow
type FollowList struct {
	APIVersion string    `json:"apiVersion" yaml:"apiVersion"`
	Kind       string    `json:"kind" yaml:"kind"`
	Accounts   []Account `json:"accounts" yaml:"accounts"`
}

type Account struct {
	Handle string `json:"handle" yaml:"handle"`
}
