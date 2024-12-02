package events

import (
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
)

type Producer interface {
	ProduceSyncEvent(ctx context.Context, evt SyncUserEvent) error
}

type SaraSyncProducer struct {
	client sarama.SyncProducer
}

func NewSaraSyncProducer(client sarama.Client) (Producer, error) {
	p, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		return nil, err
	}
	return &SaraSyncProducer{p}, nil
}

func (s *SaraSyncProducer) ProduceSyncEvent(ctx context.Context, evt SyncUserEvent) error {
	data, _ := json.Marshal(evt)
	msg := &sarama.ProducerMessage{
		Topic: "sync_user_event",
		Value: sarama.ByteEncoder(data),
	}
	_, _, err := s.client.SendMessage(msg)
	return err
}
