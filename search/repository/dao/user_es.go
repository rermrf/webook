package dao

import (
	"context"
	"encoding/json"
	"github.com/olivere/elastic/v7"
	"strconv"
	"strings"
)

const UserIndexName = "user_index"

type UserESDao struct {
	client *elastic.Client
}

func NewUserESDao(client *elastic.Client) UserDao {
	return &UserESDao{client: client}
}

func (u *UserESDao) InputUser(ctx context.Context, user User) error {
	_, err := u.client.Index().Index(UserIndexName).Id(strconv.FormatInt(user.Id, 10)).BodyJson(user).Do(ctx)
	return err
}

func (u *UserESDao) Search(ctx context.Context, keywords []string) ([]User, error) {
	// 假定上面传入的 keywords 是经过处理的
	queryString := strings.Join(keywords, " ")
	query := elastic.NewBoolQuery().Must(elastic.NewMatchQuery("nickname", queryString))
	resp, err := u.client.Search(UserIndexName).Query(query).Do(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]User, 0, len(resp.Hits.Hits))
	for _, hit := range resp.Hits.Hits {
		var user User
		err := json.Unmarshal(hit.Source, &user)
		if err != nil {
			return nil, err
		}
		res = append(res, user)
	}
	return res, nil
}
