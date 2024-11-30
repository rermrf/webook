package dao

import (
	"context"
	"encoding/json"
	"github.com/olivere/elastic/v7"
)

type TagESDao struct {
	client *elastic.Client
}

func NewTagESDao(client *elastic.Client) TagDao {
	return &TagESDao{client: client}
}

func (t *TagESDao) Search(ctx context.Context, uid int64, biz string, keywords []string) ([]int64, error) {
	query := elastic.NewBoolQuery().Must(
		elastic.NewTermsQuery("uid", uid),
		elastic.NewTermsQueryFromStrings("tags", keywords...),
		elastic.NewTermQuery("biz", biz))
	resp, err := t.client.Search(TagIndexName).Query(query).Do(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]int64, 0, len(resp.Hits.Hits))
	for _, hit := range resp.Hits.Hits {
		var elem BizTags
		err = json.Unmarshal(hit.Source, &elem)
		if err != nil {
			return nil, err
		}
		res = append(res, elem.Uid)
	}
	return res, nil
}
