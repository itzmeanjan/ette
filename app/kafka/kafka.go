package kafka

import (
	"fmt"

	kafka "github.com/segmentio/kafka-go"
)

func newKafkaWriter(kafkaURL, topic string) *kafka.Writer {
	return &kafka.Writer{
		Addr:     kafka.TCP(kafkaURL),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
}

func Connect() *kafka.Writer {
	_writer := newKafkaWriter("localhost:29092", "events")
	fmt.Println("start producing ... !!")

	return _writer
}
