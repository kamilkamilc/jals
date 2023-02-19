package store

import (
	"errors"
	"sync"

	"github.com/kamilkamilc/jals/model"
)

type MapStorage struct {
	storage map[string]*model.LinkInfo
	lock    sync.RWMutex
}

func InitializeMapStorage() *MapStorage {
	mapStorage := &MapStorage{}
	mapStorage.storage = make(map[string]*(model.LinkInfo))
	mapStorage.lock = sync.RWMutex{}
	return mapStorage
}

func (ms *MapStorage) SaveLink(link *model.Link) error {
	linkInfo := new(model.LinkInfo)
	linkInfo.Clicks = link.LinkInfo.Clicks
	linkInfo.OriginalLink = link.LinkInfo.OriginalLink
	ms.lock.Lock()
	defer ms.lock.Unlock()
	ms.storage[link.ShortLink] = linkInfo
	return nil
}

func (ms *MapStorage) RetrieveOriginalLink(shortLink string) (string, error) {
	ms.lock.RLock()
	defer ms.lock.RUnlock()
	if linkInfo, ok := ms.storage[shortLink]; !ok {
		return "", errors.New("shortLink not found")
	} else {
		return linkInfo.OriginalLink, nil
	}
}

func (ms *MapStorage) RetrieveLinkInfo(shortLink string) (*model.LinkInfo, error) {
	ms.lock.RLock()
	defer ms.lock.RUnlock()
	if li, ok := ms.storage[shortLink]; !ok {
		return nil, errors.New("shortLink not found")
	} else {
		// we don't want to return pointer to map element itself
		linkInfo := &model.LinkInfo{
			OriginalLink: li.OriginalLink,
			Clicks:       li.Clicks,
		}
		return linkInfo, nil
	}
}

func (ms *MapStorage) IncrementClicks(shortLink string) error {
	ms.lock.RLock()
	defer ms.lock.RUnlock()
	if linkInfo, ok := ms.storage[shortLink]; !ok {
		return errors.New("shortLink not found")
	} else {
		linkInfo.Clicks++
		return nil
	}
}
