package model

type LinkInfo struct {
	OriginalLink string `json:"originalLink" redis:"originalLink"`
	Clicks       int    `json:"clicks" redis:"clicks"`
}

type Link struct {
	ShortLink string   `json:"shortLink"`
	LinkInfo  LinkInfo `json:"linkInfo"`
}
