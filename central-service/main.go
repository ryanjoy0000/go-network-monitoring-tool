package main

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

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
	conf          *sarama.Config
)

func main() {
	// fetch vars from .env
	topic = os.Getenv("KAFKA_TOPIC")
	addrs_ENV := os.Getenv("KAFKA_HOST")
	addrs = strings.Split(addrs_ENV, ",")

	log.Println(".env: \n", addrs, "\n", topic)

	sarama.Logger = log.New(os.Stdout, "\t\t[sarama]", log.LstdFlags)

	conf := sarama.NewConfig()
	conf.Consumer.Return.Errors = true
	conf.Consumer.Offsets.AutoCommit.Enable = true
	conf.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	conf.ClientID = "network-metrics"

	partitionConsumer, consumer := createConsumer()

	startConsuming(partitionConsumer, consumer)
}

func createConsumer() (sarama.PartitionConsumer, sarama.Consumer) {
	// create kafka consumer
	consumer, err := sarama.NewConsumer(addrs, conf)
	if err != nil {
		log.Panicln("Error while creating consumer", err)
	}
	log.Println("master consumer created...")

	// creata a PartitionConsumer
	partitionConsumer, err := consumer.ConsumePartition(
		topic,
		0,
		sarama.OffsetNewest,
	)
	if err != nil {
		log.Panicln("Error while creating partition consumer", err)
	}
	log.Println("partition consumer created...")

	return partitionConsumer, consumer
}

func handleConsumerClose(c sarama.PartitionConsumer, mainC sarama.Consumer) {
	err := c.Close()
	if err != nil {
		log.Panicln("Error while closing the partitionConsumer", err)
	}
	log.Println("closing partitionConsumer....")

	err = mainC.Close()
	if err != nil {
		log.Panicln("Error while closing the consumer", err)
	}
	log.Println("closing consumer....")
}

func startConsuming(c sarama.PartitionConsumer, mainC sarama.Consumer) {
	// Trap SIGINT to trigger a graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	consumedCount := 0

	canRun := true
	log.Println("ready to consume msgs...")
	for canRun {
		select {
		case msg := <-c.Messages():
			log.Println("-------------------------")
			log.Println("RECEIVED =====> ", string(msg.Value))
			log.Println("FROM => ", "Topic: ", msg.Topic, "| Partition:", msg.Partition, " | Offset:", msg.Offset)
			consumedCount++
			log.Println("Consumed msgs till now: ", consumedCount)
			log.Println("-------------------------")

		// handle error
		case err := <-c.Errors():
			log.Println("Error while consuming msg", err)
			canRun = false

			// Handle interruption and exit
		case <-signalChan:
			log.Println("Interruption.. Consumer exiting...")

		}
	}

	log.Println("Total consumed msgs: ", consumedCount)

	// close consumer properly in the end
	defer handleConsumerClose(c, mainC)
}
