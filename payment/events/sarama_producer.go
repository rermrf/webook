package events

import (
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
)

type SaramaProducer struct {
	producer sarama.SyncProducer
}

func NewSaramaProducer(client sarama.Client) (Producer, error) {
	p, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		return nil, err
	}
	return &SaramaProducer{p}, nil
}

func (s *SaramaProducer) ProducePaymentEvent(ctx context.Context, event PaymentEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, _, err = s.producer.SendMessage(&sarama.ProducerMessage{
		Key:   sarama.StringEncoder(event.BizTradeNO),
		Topic: event.Topic(),
		Value: sarama.ByteEncoder(data),
	})
	return err
}
