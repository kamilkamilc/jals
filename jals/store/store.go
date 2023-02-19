package store

import "github.com/kamilkamilc/jals/model"

type Storage interface {
	SaveLink(link *model.Link) error
	RetrieveOriginalLink(shortLink string) (string, error)
	RetrieveLinkInfo(shortLink string) (*model.LinkInfo, error)
	IncrementClicks(shortLink string) error
}
