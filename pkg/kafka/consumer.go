package kafka

import (
	"context"

	kafkago "github.com/segmentio/kafka-go"
)

type Message struct {
	Key     []byte
	Value   []byte
	Headers map[string]string
}

type Handler func(ctx context.Context, msg Message) error

type Consumer struct {
	reader *kafkago.Reader
}

type ConsumerConfig struct {
	Brokers []string
	Topic   string
	GroupID string
}

func NewConsumer(cfg ConsumerConfig) *Consumer {
	return &Consumer{
		reader: kafkago.NewReader(kafkago.ReaderConfig{
			Brokers: cfg.Brokers,
			Topic:   cfg.Topic,
			GroupID: cfg.GroupID,
		}),
	}
}

func (c *Consumer) Run(ctx context.Context, handler Handler) error {
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			return err
		}

		headers := make(map[string]string, len(msg.Headers))
		for _, header := range msg.Headers {
			headers[header.Key] = string(header.Value)
		}

		err = handler(ctx, Message{
			Key:     msg.Key,
			Value:   msg.Value,
			Headers: headers,
		})
		if err != nil {
			return err
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			return err
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
