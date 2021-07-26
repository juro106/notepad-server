package models

type Content struct {
	User    string   `json:"user"`
	Title   string   `json:"title"`
	Slug    string   `json:"slug"`
	Tags    []string `json:"tags"`
	Content string   `json:"content"`
}
