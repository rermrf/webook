package dao

import (
	"context"
	"github.com/olivere/elastic/v7"
)

type AnyESDao struct {
	client *elastic.Client
}

func NewAnyESDao(client *elastic.Client) AnyDao {
	return &AnyESDao{client: client}
}

func (a *AnyESDao) Input(ctx context.Context, index, docId, data string) error {
	_, err := a.client.Index().Index(index).Id(docId).BodyString(data).Do(ctx)
	return err
}
