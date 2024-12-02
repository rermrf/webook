package intergration

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
	searchv1 "webook/api/proto/gen/search/v1"
	"webook/search/grpc"
	"webook/search/intergration/startup"
)

type SearchTestSuite struct {
	suite.Suite
	searchSvc *grpc.SearchServiceServer
	syncSvc   *grpc.SyncServiceServer
}

func (s *SearchTestSuite) SetupTest() {
	s.searchSvc = startup.InitSearchServer()
	s.syncSvc = startup.InitSyncServer()
}

func (s *SearchTestSuite) TestSearch() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	data, err := json.Marshal(BizTags{
		Uid:   577,
		Biz:   "article",
		BizId: 123,
		Tags:  []string{"Jerry"},
	})
	require.NoError(s.T(), err)
	_, err = s.syncSvc.InputUser(ctx, &searchv1.InputUserRequest{
		User: &searchv1.User{
			Id:       123,
			Nickname: "Tom",
		},
	})
	require.NoError(s.T(), err)
	_, err = s.syncSvc.InputAny(ctx, &searchv1.InputAnyRequest{
		IndexName: "tags_index",
		DocId:     "abc",
		Data:      string(data),
	})
	require.NoError(s.T(), err)
	_, err = s.syncSvc.InputArticle(ctx, &searchv1.InputArticleRequest{
		Article: &searchv1.Article{
			Id:     123,
			Title:  "This is a title",
			Status: 2,
		},
	})
	require.NoError(s.T(), err)
	_, err = s.syncSvc.InputArticle(ctx, &searchv1.InputArticleRequest{
		Article: &searchv1.Article{
			Id:      124,
			Content: "This is a content",
			Status:  2,
		},
	})
	require.NoError(s.T(), err)
	resp, err := s.searchSvc.Search(ctx, &searchv1.SearchRequest{
		Expression: "title content Tom Jerry",
		Uid:        1001,
	})
	s.T().Log(resp)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 1, len(resp.User.Users))
	assert.Equal(s.T(), 2, len(resp.Article.Articles))
}

type BizTags struct {
	Uid   int64    `json:"uid"`
	Biz   string   `json:"biz"`
	BizId int64    `json:"biz_id"`
	Tags  []string `json:"tags"`
}

func TestSearchService(t *testing.T) {
	suite.Run(t, new(SearchTestSuite))
}
