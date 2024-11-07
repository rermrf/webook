package events

import (
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
)

type SaramaProducer struct {
	p     sarama.SyncProducer
	topic string
}

func NewSaramaProducer(p sarama.SyncProducer, topic string) *SaramaProducer {
	return &SaramaProducer{p: p, topic: topic}
}

func (p *SaramaProducer) ProduceInconsistentEvent(ctx context.Context, event InconsistentEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, _, err = p.p.SendMessage(&sarama.ProducerMessage{
		Topic: p.topic,
		Value: sarama.ByteEncoder(data),
	})
	return err
}
