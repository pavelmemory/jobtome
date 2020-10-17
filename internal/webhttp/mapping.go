package webhttp

import (
	"github.com/pavelmemory/jobtome/internal/shorten"
)

type CreateShortenReq struct {
	URL string `json:"url"`
}

type GetShortenResp struct {
	ID   int64  `json:"id"`
	URL  string `json:"url"`
	Hash string `json:"hash"`
}

type ListShortenResp []GetShortenResp

type Mapper struct{}

func (m Mapper) createShortenReq2Entity(req CreateShortenReq) shorten.Entity {
	return shorten.Entity{URL: req.URL}
}

func (Mapper) entity2GetShortenResp(entity shorten.Entity) GetShortenResp {
	return (GetShortenResp)(entity)
}

func (m Mapper) entities2ListShortenResp(entities []shorten.Entity) ListShortenResp {
	res := make(ListShortenResp, len(entities))
	for i, entity := range entities {
		res[i] = m.entity2GetShortenResp(entity)
	}

	return res
}
