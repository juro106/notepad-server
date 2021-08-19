package models

type Content struct {
	User       string   `json:"user"`
	Title      string   `json:"title"`
	Slug       string   `json:"slug"`
	Tags       []string `json:"tags"`
	Content    string   `json:"content"`
	Created_at string   `json:"created_at"`
	Updated_at string   `json:"updated_at"`
	Project    string   `json:"project"`
	Image      string   `json:"image"`
}
