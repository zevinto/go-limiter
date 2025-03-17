package mq_kafka

import (
	"fmt"
	"github.com/segmentio/kafka-go"
	"log"
	"time"
)

func StartProducer() {
	writer := &kafka.Writer{
		Addr:  kafka.TCP("localhost:9092"),
		Topic: "request_queue",
	}

	for i := 0; i < 10; i++ {
		err := writer.WriteMessages(nil, kafka.Message{Value: []byte(fmt.Sprintf("Request %d", i))})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Produced Request:", i)
		time.Sleep(100 * time.Millisecond) // 控制生产速率
	}
}
