package news

import (
	"context"

	"github.com/barthr/newsapi"
)

var _ INews = (*mockNewsClient)(nil)

type mockNewsClient struct {
	news []newsapi.Article
}

func NewMockNewsClient(news []newsapi.Article) INews {
	return &mockNewsClient{
		news: news,
	}
}

func (m *mockNewsClient) GetEverything(ctx context.Context, params *newsapi.EverythingParameters) (*newsapi.ArticleResponse, error) {
	ar := &newsapi.ArticleResponse{
		Status:       "ok",
		TotalResults: len(m.news),
		Articles:     make([]newsapi.Article, 0),
	}

	ar.Articles = append(ar.Articles, m.news...)

	return ar, nil
}
