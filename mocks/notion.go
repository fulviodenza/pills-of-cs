package mocks

import (
	"context"
	"github.com/jomei/notionapi"
)

type NotionDatabaseServiceMock struct {
	QueryVal *notionapi.DatabaseQueryResponse
	QueryErr error
}

var _ notionapi.DatabaseService = (*NotionDatabaseServiceMock)(nil)

func (n NotionDatabaseServiceMock) Get(ctx context.Context, id notionapi.DatabaseID) (*notionapi.Database, error) {
	//TODO implement me
	panic("implement me")
}

func (n NotionDatabaseServiceMock) Query(ctx context.Context, id notionapi.DatabaseID, request *notionapi.DatabaseQueryRequest) (*notionapi.DatabaseQueryResponse, error) {
	return n.QueryVal, n.QueryErr
}

func (n NotionDatabaseServiceMock) Update(ctx context.Context, id notionapi.DatabaseID, request *notionapi.DatabaseUpdateRequest) (*notionapi.Database, error) {
	//TODO implement me
	panic("implement me")
}

func (n NotionDatabaseServiceMock) Create(ctx context.Context, request *notionapi.DatabaseCreateRequest) (*notionapi.Database, error) {
	//TODO implement me
	panic("implement me")
}
