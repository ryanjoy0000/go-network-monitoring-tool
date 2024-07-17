package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/shirou/gopsutil/v4/net"
)

type NetworkMetrics struct {
	Device    string  `json:"device"`
	Timestamp int64   `json:"timestamp"`
	Latency   float64 `json:"latency"`
	Bandwidth float64 `json:"bandwidth"`
}

func Start_data_collector() {
	latencyChan := make(chan float64)

	// kafka config
	config := &kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
	}

	// create Kafka producer
	producer, err := kafka.NewProducer(config)
	if err != nil {
		log.Panic("Failed to create producer", err)
	}
	defer producer.Close() // close producer in the end

	log.Println("producer created...")
	canRun := true
	for canRun {
		// get network I/O statistics for every network interface installed on the system
		netStatList, err := net.IOCounters(false)
		if err != nil {
			log.Panicln("Error while collecting network metrics: ", err)
			canRun = false
		} else {
			for _, val := range netStatList {

				go calcLatency("google.com", latencyChan)

				metrics := NetworkMetrics{
					Timestamp: time.Now().Unix(),
					Device:    val.Name,
					Latency:   <-latencyChan,
					Bandwidth: float64(val.BytesRecv+val.BytesSent) / 1024, // KBps
				}

				fmt.Println("metrics: ", metrics)

				// serialize metrics to JSON
				metricsJson, err := json.Marshal(metrics)
				if err != nil {
					log.Panic("Error while converting metrics to JSON: ", err)
				}

				// kafka topic
				topic := "network-metrics"

				// kafka message
				msg := &kafka.Message{
					TopicPartition: kafka.TopicPartition{
						Topic: &topic,
					},
					Value: metricsJson,
				}

				// produce the msg to Kafka
				producer.Produce(msg, nil)

			}
			// wait to get confirmation of message deliveries
			producer.Flush(5 * 1000)
		}
	}

	// Wait and check for successful deliveries of messages to Kafka
	go func() {
		for ev := range producer.Events() {
			switch event := ev.(type) {
			case *kafka.Message:
				if event.TopicPartition.Error != nil {
					log.Println("Could not deliver msg: ", event.TopicPartition)
				} else {
					log.Println("Msg delivered to: ", event.TopicPartition)
				}
			}
		}
	}()
}

func calcLatency(reliableHost string, latencyChan chan float64) {
	// run ping to a reliable host
	pingResult, err := exec.Command("ping", "-c", "4", reliableHost).Output()
	if err != nil {
		log.Panicln("Error while pinging host:", err)
	}

	// convert result to string
	pingStr := string(pingResult)

	log.Println("pingStr: ", pingStr)

	// Find the line with the average latency
	r := regexp.MustCompile(`round-trip min/avg/max/stddev = [\d\.]+/([\d\.]+)/[\d\.]+/[\d\.]+ ms`)
	log.Println("r:", r)
	matches := r.FindStringSubmatch(pingStr)
	log.Println("matches:", matches)
	if len(matches) < 2 {
		log.Panicln("Failed to parse ping output: ", pingStr)
	}

	// Convert the average latency to float64
	avgLatency, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		log.Panicln("Failed to convert latency to float: v", err)
	}
	log.Println("avgLatency", avgLatency)

	latencyChan <- avgLatency
}
