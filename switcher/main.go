package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
	"golang.org/x/exp/maps"
)

var switchWriter *kafka.Writer

func listTopics() ([]string, error) {
	conn, err := kafka.Dial("tcp", getBrokerURLs()[0])
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

func main() {
	fmt.Println("Starting Kafka Producer")

	// while kafka has not started, retry
	for {
		_, err := listTopics()
		if err != nil {
			log.Printf("error connecting to kafka, error: %v", err)
			time.Sleep(time.Duration(2) * time.Second)
		}
		if err == nil {
			break
		}
	}
	switchWriter = getKafkaWriter("switch")
	writeLoop()
}

func getKafkaWriter(topic string) *kafka.Writer {
	return &kafka.Writer{
		Addr:            kafka.TCP(getBrokerURLs()...),
		Topic:           topic,
		Balancer:        &kafka.LeastBytes{},
		WriteBackoffMax: time.Duration(500) * time.Millisecond,
		BatchTimeout:    time.Duration(500) * time.Millisecond,
	}
}

func getBrokerURLs() []string {
	BrokerURLs := strings.Split(os.Getenv("MSK_BROKERS"), ",")
	for i := range BrokerURLs {
		BrokerURLs[i] = strings.TrimSpace(BrokerURLs[i])
	}
	fmt.Println("Brokers:", BrokerURLs)
	return BrokerURLs
}

func writeLoop() {
	fmt.Println("Starting Write Loop")
	messages := []string{"Counter", "Text", "Time", "Counter", "Text", "Time", "Counter", "Text", "Time", "Dump"}
	orgs := []string{"1", "2", "3", "4"}

	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)

	for {
		org := orgs[r.Intn(len(orgs))]
		msg := messages[r.Intn(len(messages))]
		writeMessage(org, msg)

		log.Printf("Wrote message: %s org: %s", msg, org)

		sleepTime := time.Duration(r.Intn(5)+1) * time.Second
		time.Sleep(sleepTime)
	}
}

func writeMessage(org string, msg string) {
	msgJson := fmt.Sprintf(`{"org": "%s", "msg": "%s"}`, org, msg)
	err := switchWriter.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte(time.Now().String()),
			Value: []byte(msgJson),
		},
	)
	if err != nil {
		fmt.Println("Error while writing to kafka", err)
	}
}
