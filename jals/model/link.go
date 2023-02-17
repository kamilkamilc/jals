package model

type Link struct {
	ShortLink    string `json:"shortLink"`
	OriginalLink string `json:"originalLink"`
	Clicks       int    `json:"clicks"`
}
