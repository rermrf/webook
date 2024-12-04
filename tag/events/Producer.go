package events

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
)

type Producer interface {
	ProduceSyncEvent(ctx context.Context, data BizTags) error
}

type SaramaSyncProducer struct {
	client sarama.SyncProducer
}

func NewSaramaSyncProducer(client sarama.SyncProducer) *SaramaSyncProducer {
	return &SaramaSyncProducer{client: client}
}

func (p *SaramaSyncProducer) ProduceSyncEvent(ctx context.Context, tags BizTags) error {
	val, err := json.Marshal(tags)
	if err != nil {
		return err
	}
	evt := SyncDataEvent{
		IndexName: "tags_index",
		// 构成一个唯一的 doc id
		// 要确保后面打了新标签的时候，搜索那边也会有对应的修改
		DocId: fmt.Sprintf("%d_%s_%d", tags.Uid, tags.Biz, tags.BizId),
		Data:  string(val),
	}
	val, err = json.Marshal(evt)
	if err != nil {
		return err
	}
	_, _, err = p.client.SendMessage(&sarama.ProducerMessage{
		Topic: "sync_search_data",
		Value: sarama.StringEncoder(val),
	})
	return err
}

type BizTags struct {
	Tags  []string `json:"tags"`
	Biz   string   `json:"biz"`
	BizId int64    `json:"biz_id"`
	Uid   int64    `json:"uid"`
}
