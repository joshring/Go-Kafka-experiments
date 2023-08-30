package kafka_utils

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
)

func GetReader(
	topic string,
	startOffset int64,
) *kafka.Reader {

	reader := kafka.NewReader(
		kafka.ReaderConfig{
			Brokers: GetBrokerURLs(),
			Topic:   topic,
		},
	)
	if os.Getenv("DEV_ENV") == "TRUE" {
		// Removes "catch up period" at the start of process start
		// making development easier with auto-recompilation enabled, disable in prod
		// leads to data loss for events before this process started, so unsuitable for production
		reader.SetOffsetAt(context.Background(), time.Now())
	}

	return reader
}

func ReadMessage(
	kafkaReader *kafka.Reader,
) (kafka.Message, error) {

	// Create a context with a timeout, limiting wait period
	ctxTimeout, cancel := context.WithTimeout(context.Background(), time.Duration(1)*time.Second)
	defer cancel()

	message, err := kafkaReader.ReadMessage(ctxTimeout)
	if err != nil {
		return kafka.Message{}, fmt.Errorf("unable to read message from kafka queue")
	}

	return message, nil
}
