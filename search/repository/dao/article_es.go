package dao

import (
	"context"
	"encoding/json"
	"github.com/olivere/elastic/v7"
	"strconv"
	"strings"
)

const ArticleIndexName = "article_index"
const TagIndexName = "tags_index"

type ArticleESDao struct {
	client *elastic.Client
}

func NewArticleESDao(client *elastic.Client) ArticleDao {
	return &ArticleESDao{client: client}
}

func (a *ArticleESDao) InputArticle(ctx context.Context, article Article) error {
	_, err := a.client.Index().
		Index(ArticleIndexName).
		// 为社么指定 ID？
		// 确保后面文章更新的时候，我们这里产生类似的两条数据，而是更新了数据
		Id(strconv.FormatInt(article.Id, 10)).
		BodyJson(article).Do(ctx)
	return err
}

func (a *ArticleESDao) Search(ctx context.Context, tagArtIds []int64, keywords []string) ([]Article, error) {
	queryString := strings.Join(keywords, " ")
	ids := make([]any, 0)
	for _, id := range tagArtIds {
		ids = append(ids, strconv.FormatInt(id, 10))
	}
	query := elastic.NewBoolQuery().Must(
		elastic.NewBoolQuery().Should(
			// 给予更高权重
			elastic.NewTermsQuery("id", ids...).Boost(2),
			elastic.NewMatchQuery("title", queryString),
			elastic.NewMatchQuery("content", queryString)),
		elastic.NewTermQuery("status", 2),
	)
	resp, err := a.client.Search(ArticleIndexName).Query(query).Do(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]Article, 0, len(resp.Hits.Hits))
	for _, hit := range resp.Hits.Hits {
		var elem Article
		err = json.Unmarshal(hit.Source, &elem)
		res = append(res, elem)
	}
	return res, nil
}
