package mq_kafka

import (
	"fmt"
	"github.com/segmentio/kafka-go"
	"log"
	"time"
)

func StartConsumer() {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "request_queue",
		GroupID: "consumer-group",
	})

	for {
		msg, err := reader.ReadMessage(nil)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Consumed:", string(msg.Value))
		time.Sleep(1 * time.Second) // 限制消费速率
	}
}
