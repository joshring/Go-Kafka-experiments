package main

import (
	"encoding/json"
	"fmt"
	"kafka_manager/kafka_utils"
	"log"
	"os"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
)

type MessageValue struct {
	Organisation string `json:"org"`
	Message      string `json:"msg"`
}

func processDump(
	readerSwitch *kafka.Reader,
	currentOffset int64,
) {
	readerSwitch.SetOffset(kafka.FirstOffset)

	dumpMessageValueBuffer := []MessageValue{}
	dumpMessageValue := MessageValue{}

	for {
		message, err := kafka_utils.ReadMessage(readerSwitch)
		if err != nil {
			break
		}
		if message.Offset > currentOffset {
			break
		}

		err = json.Unmarshal(message.Value, &dumpMessageValue)
		if err != nil {
			log.Printf("unable to unmarshal message.Value: %s error: %v", message.Value, err)
			continue
		}
		// the instructions classify outputs as: (Counter, Time, Text)
		// so exclude "Dump" messages from the Dumped json file
		if dumpMessageValue.Message == "Dump" {
			continue
		}
		dumpMessageValueBuffer = append(dumpMessageValueBuffer, dumpMessageValue)
	}

	jsonBytesData, err := json.MarshalIndent(
		dumpMessageValueBuffer,
		"",
		"\t",
	)
	if err != nil {
		log.Printf("unable to marshal data to bytes for json export to file")
		return
	}

	err = os.WriteFile(
		fmt.Sprintf("../output_dumps/%s.json", time.Now().Format("2006-01-02T15:04:05Z")),
		jsonBytesData,
		os.ModePerm,
	)
	if err != nil {
		log.Printf("unable to save json bytes data to file")
		return
	}

	// reset the reader's offset
	readerSwitch.SetOffset(currentOffset)

}

func processSwitchMessage(
	message kafka.Message,
	readerSwitch *kafka.Reader, // for the "Dump" case
	writerText *kafka.Writer,
	writerTime *kafka.Writer,
	writerCounter *kafka.Writer,
) {

	messageValue := MessageValue{}

	err := json.Unmarshal(message.Value, &messageValue)
	if err != nil {
		log.Printf("unable to unmarshal message.Value: %s error: %v", message.Value, err)
		return
	}

	log.Printf("The message is: %+v", messageValue.Message)

	switch messageValue.Message {
	case "Text":
		kafka_utils.WriteMessage(
			messageValue.Organisation,
			messageValue.Message,
			writerText,
		)
		return

	case "Time":
		kafka_utils.WriteMessage(
			messageValue.Organisation,
			messageValue.Message,
			writerTime,
		)
		return

	case "Counter":
		kafka_utils.WriteMessage(
			messageValue.Organisation,
			messageValue.Message,
			writerCounter,
		)
		return

	case "Dump":
		currentOffset := readerSwitch.Offset()
		processDump(
			readerSwitch,
			currentOffset,
		)
		return

	default:
		log.Printf("unrecognised message: %s", messageValue.Message)
	}

}

func readLoop(
	readerSwitch *kafka.Reader,
	writerText *kafka.Writer,
	writerTime *kafka.Writer,
	writerCounter *kafka.Writer,
) {

	// monitor where we are in the cycle,
	//	buffering messages every second,
	// 	processing messages every 3 seconds
	counter := 0
	messageBuffer := []kafka.Message{}

	for {
		time.Sleep(time.Second)

		messageSwitch, err := kafka_utils.ReadMessage(readerSwitch)
		if err == nil {
			messageBuffer = append(messageBuffer, messageSwitch)
		}

		// every 3 seconds process messageBuffer in parallel
		if counter%3 == 0 {

			log.Printf("messages read: %d", len(messageBuffer))

			wg := sync.WaitGroup{}

			for _, message := range messageBuffer {

				// process the buffered messages in parallel
				wg.Add(1)
				go func(
					message kafka.Message,
					readerSwitch *kafka.Reader,
					writerText *kafka.Writer,
					writerTime *kafka.Writer,
					writerCounter *kafka.Writer,
				) {
					defer wg.Done()
					processSwitchMessage(
						message,
						readerSwitch,
						writerText,
						writerTime,
						writerCounter,
					)
				}(
					message,
					readerSwitch,
					writerText,
					writerTime,
					writerCounter,
				)
			}

			// reset buffer after processing
			messageBuffer = []kafka.Message{}

		}
		counter++

	}

}

func main() {

	fmt.Println("Starting Kafka Message Manager")
	// while kafka has not started, retry
	for {
		err := kafka_utils.CreateTopics([]string{"Counter", "Time", "Text"})
		if err != nil {
			log.Printf("error creating required kafka topic(s), error: %v", err)
			time.Sleep(time.Duration(2) * time.Second)
		}
		if err == nil {
			break
		}
	}

	readerSwitch := kafka_utils.GetReader("switch", 0)
	defer readerSwitch.Close()

	writerText := kafka_utils.GetWriter("Text")
	writerTime := kafka_utils.GetWriter("Time")
	writerCounter := kafka_utils.GetWriter("Counter")

	readLoop(
		readerSwitch,
		writerText,
		writerTime,
		writerCounter,
	)
}
