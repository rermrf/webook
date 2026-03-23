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
		// 以 biz_bizId 作为唯一 doc id（不再按用户隔离）
		DocId: fmt.Sprintf("%s_%d", tags.Biz, tags.BizId),
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
}

