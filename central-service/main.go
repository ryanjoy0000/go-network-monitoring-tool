package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/IBM/sarama"
)

const (
	ENV = "../.env"
)

type NetworkMetrics struct {
	Device    string  `json:"device"`
	Timestamp int64   `json:"timestamp"`
	Latency   float64 `json:"latency"`
	Bandwidth float64 `json:"bandwidth"`
}

var (
	topic         string
	addrs         []string
	reliable_host string
)

func main() {
	// fetch vars from .env
	topic = os.Getenv("KAFKA_TOPIC")
	addrs = []string{os.Getenv("KAFKA_HOST")}

	log.Println(".env: \n", addrs, "\n", topic)

	// create kafka consumer
	consumer, err := sarama.NewConsumer(addrs, nil)
	if err != nil {
		log.Panicln("Error while creating conusmer", err)
	}
	// creata a PartitionConsumer
	partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
	if err != nil {
		log.Panicln("Error while creating partition conusmer", err)
	}
	log.Println("kafka consumer ready...")

	// close consumer properly in the end
	defer handleConsumerClose(consumer)

	startConsuming(partitionConsumer)
}

func handleConsumerClose(c sarama.Consumer) {
	err := c.Close()
	if err != nil {
		log.Panicln("Error while closing the consumer", err)
	}
	log.Println("closing consumer....")
}

func startConsuming(c sarama.PartitionConsumer) {
	// Trap SIGINT to trigger a graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	consumedCount := 0

ConsumerLoop:
	for {
		select {
		case msg := <-sarama.PartitionConsumer.Messages(c):
			log.Println("============> received: ", string(msg.Value))
			consumedCount++

		// Handle interruption and exit
		case <-signalChan:
			log.Println("Interruption.. Consumer exiting...")
			break ConsumerLoop
		}
	}
	log.Println("Total consumed msgs: ", consumedCount)
}
