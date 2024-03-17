package news

import (
	"context"

	"github.com/barthr/newsapi"
)

var _ INews = (*NewsAdapter)(nil)

type INews interface {
	GetEverything(ctx context.Context, params *newsapi.EverythingParameters) (*newsapi.ArticleResponse, error)
}

type NewsAdapter struct {
	newsapi.Client
}

func NewNewsAdapter(client newsapi.Client) INews {
	return &NewsAdapter{
		Client: client,
	}
}

func (n *NewsAdapter) GetEverything(ctx context.Context, params *newsapi.EverythingParameters) (*newsapi.ArticleResponse, error) {
	return n.Client.GetEverything(ctx, params)
}
