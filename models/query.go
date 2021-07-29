package models

type Query struct {
	Slug string   `json:"slug"`
	Uid  string   `json:"uid"`
	Tags []string `json:"tags"`
}
