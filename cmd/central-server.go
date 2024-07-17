package main

import (
	"log"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

func Start_central_server() {
	// Kafka config - consumer
	config := &kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
		"group.id":          "network-monitoring",
		"auto.offset.reset": "earliest",
	}

	// Kafka consumer
	consumer, err := kafka.NewConsumer(config)
	if err != nil {
		log.Panicln(err)
	}
	defer consumer.Close()
	log.Println("consumer created...")

	// create kafka topic
	topics := []string{"network-metrics"}

	// Subscribe to the kafka topic
	consumer.SubscribeTopics(topics, nil)

	canRun := true
	for canRun {
		// Poll for a message
		msg, err := consumer.ReadMessage(time.Second)
		if err != nil && !err.(kafka.Error).IsTimeout() {
			log.Panic(err)
			canRun = false
		} else if msg != nil {
			// received a message
			log.Println("==========> received: ", string(msg.Value))
		}
	}
}
