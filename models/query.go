package models

type Query struct {
	Slug    string   `json:"slug"`
	Uid     string   `json:"uid"`
	Project string   `json:"project"`
	Tags    []string `json:"tags"`
}
