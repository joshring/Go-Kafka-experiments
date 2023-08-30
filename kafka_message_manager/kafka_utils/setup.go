package kafka_utils

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/segmentio/kafka-go"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

func GetBrokerURLs() []string {
	BrokerURLs := strings.Split(os.Getenv("MSK_BROKERS"), ",")
	for i := range BrokerURLs {
		BrokerURLs[i] = strings.TrimSpace(BrokerURLs[i])
	}
	fmt.Println("Brokers:", BrokerURLs)
	return BrokerURLs
}

// used to test the kafka connection is working on container startup
// and to verify topics were created successfully
func ListTopics() ([]string, error) {
	conn, err := kafka.Dial("tcp", GetBrokerURLs()[0])
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	partitionSlice, err := conn.ReadPartitions()
	if err != nil {
		return nil, err
	}

	partitionTopics := map[string]struct{}{}

	// insert empty struct since we're only interested in the unique keys, the topics
	for _, partition := range partitionSlice {
		partitionTopics[partition.Topic] = struct{}{}
	}

	return maps.Keys(partitionTopics), nil

}

// creates the topic since KAFKA_AUTO_CREATE_TOPICS_ENABLE='true'
func CreateTopics(kafkaTopics []string) error {

	for _, topic := range kafkaTopics {
		_, err := kafka.DialLeader(
			context.Background(),
			"tcp",
			GetBrokerURLs()[0],
			topic,
			0,
		)
		if err != nil {
			return err
		}
	}

	// verify topics created
	kafkaTopicsRead, err := ListTopics()
	if err != nil {
		return err
	}
	log.Printf("List of kafka topics: %+v", kafkaTopicsRead)

	for _, topic := range kafkaTopics {
		if !slices.Contains(kafkaTopicsRead, topic) {
			return fmt.Errorf("unable to create topic: %s", topic)
		}
	}

	return nil
}
