package main

import (
	"log"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

func Start_central_server() {
	// Kafka config - consumer
	config := &kafka.ConfigMap{
		"bootstrap.servers": "localhost:9093",
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

// docker run -d --name kafka -p 9092:9092 -p 9093:9093 \
//   -e KAFKA_CFG_NODE_ID=0 \
//   -e KAFKA_CFG_PROCESS_ROLES=broker,controller \
//   -e KAFKA_CFG_CONTROLLER_QUORUM_VOTERS=0@localhost:9093 \
//   -e KAFKA_CFG_LISTENERS=PLAINTEXT://:9092,CONTROLLER://:9093 \
//   -e KAFKA_CFG_LISTENER_SECURITY_PROTOCOL_MAP=PLAINTEXT:PLAINTEXT,CONTROLLER:PLAINTEXT \
//   -e KAFKA_CFG_INTER_BROKER_LISTENER_NAME=PLAINTEXT \
//   -e KAFKA_CFG_CONTROLLER_LISTENER_NAMES=CONTROLLER \
//   -e KAFKA_CFG_ADVERTISED_LISTENERS=PLAINTEXT://localhost:9092 \
//   -e ALLOW_PLAINTEXT_LISTENER=yes \
//   -v kafka_data:/bitnami/kafka \
//   bitnami/kafka:latest
