package startup

import (
	"github.com/olivere/elastic/v7"
	"time"
	"webook/search/repository/dao"
)

func InitESClient() *elastic.Client {
	const timeout = 100 * time.Second
	opts := []elastic.ClientOptionFunc{
		elastic.SetURL("http://localhost:9200/"),
		elastic.SetHealthcheckTimeoutStartup(timeout),
		elastic.SetSniff(false),
	}
	client, err := elastic.NewClient(opts...)
	if err != nil {
		panic(err)
	}
	err = dao.InitES(client)
	if err != nil {
		panic(err)
	}
	return client
}
