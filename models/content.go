package models

type Content struct {
	User    string   `json:"user"`
	Title   string   `json:"title"`
	Tags    []string `json:"tags"`
	Content string   `json:"content"`
}
