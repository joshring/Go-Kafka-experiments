package kafka_utils

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

func GetWriter(
	topic string,
) *kafka.Writer {
	return &kafka.Writer{
		Addr:            kafka.TCP(GetBrokerURLs()...),
		Topic:           topic,
		Balancer:        &kafka.LeastBytes{},
		WriteBackoffMin: time.Duration(50) * time.Millisecond,
		WriteBackoffMax: time.Duration(500) * time.Millisecond,
		BatchTimeout:    time.Duration(1000) * time.Millisecond,
	}
}

func WriteMessage(
	org string,
	msg string,
	writer *kafka.Writer,
) {
	msgJson := fmt.Sprintf(`{"org": "%s", "msg": "%s"}`, org, msg)
	err := writer.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte(time.Now().String()),
			Value: []byte(msgJson),
		},
	)
	if err != nil {
		fmt.Println("Error while writing to kafka", err)
	}
}
