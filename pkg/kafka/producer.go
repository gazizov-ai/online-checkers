package kafka

import (
	"context"

	kafkago "github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafkago.Writer
}

type ProducerConfig struct {
	Brokers []string
	Topic   string
}

func NewProducer(cfg ProducerConfig) *Producer {
	return &Producer{
		writer: &kafkago.Writer{
			Addr:     kafkago.TCP(cfg.Brokers...),
			Topic:    cfg.Topic,
			Balancer: &kafkago.LeastBytes{},
		},
	}
}

func (p *Producer) Publish(
	ctx context.Context,
	key []byte,
	value []byte,
	headers map[string]string,
) error {
	msg := kafkago.Message{
		Key:   key,
		Value: value,
	}

	if len(headers) > 0 {
		msg.Headers = make([]kafkago.Header, 0, len(headers))
		for k, v := range headers {
			msg.Headers = append(msg.Headers, kafkago.Header{
				Key:   k,
				Value: []byte(v),
			})
		}
	}

	return p.writer.WriteMessages(ctx, msg)
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
