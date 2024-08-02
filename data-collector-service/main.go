package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/IBM/sarama"
	"github.com/shirou/gopsutil/v4/net"
)

const (
	DELAY_SEC = 3
)

type NetworkMetrics struct {
	Device          string  `json:"device"`
	Timestamp       int64   `json:"timestamp"`
	Latency         float64 `json:"latency"`
	BytesSent       int64   `json:"bytes_sent"`
	BytesReceived   int64   `json:"bytes_received"`
	PacketsSent     int64   `json:"packets_sent"`
	PacketsReceived int64   `json:"packets_received"`
}

var (
	topic         string
	addrs         []string
	reliable_host string
)

func main() {
	// fetch vars from .env
	topic = os.Getenv("KAFKA_TOPIC")
	addrs_ENV := os.Getenv("KAFKA_HOST")
	addrs = strings.Split(addrs_ENV, ",")
	reliable_host = os.Getenv("RELIABLE_HOST")

	log.Println(".env: \n", addrs, "\n", topic, "\n", reliable_host)

	// create new topic with cluster admin
	conf := sarama.NewConfig()
	conf.Producer.Return.Successes = true
	clusterAdmin, err := sarama.NewClusterAdmin(addrs, conf)
	if err != nil {
		log.Panicln("Error while creating cluster admin", err)
	}
	defer func() { clusterAdmin.Close() }()

	// create kafka producer
	producer, err := sarama.NewAsyncProducer(addrs, conf)
	if err != nil {
		log.Panicln("Error while creating producer", err)
	}
	log.Println("kafka producer ready...")

	// close producer properly in the end
	defer handleProducerClose(producer)

	// start collection of network metrics
	startDataCollection(producer)
}

func handleProducerClose(p sarama.AsyncProducer) {
	err := p.Close()
	if err != nil {
		log.Panicln("Error while closing the producer", err)
	}
	log.Println("closing producer....")
}

func startDataCollection(producer sarama.AsyncProducer) {
	latencyChan := make(chan float64)

	canRun := true
	for canRun {
		// --------------------------------------------------------------------------------------
		// get network I/O statistics for every network interface installed on the system
		netStatList, err := net.IOCounters(false)
		if err != nil {
			log.Panicln("Error while collecting network metrics: ", err)
			canRun = false
		} else {
			// log.Println("netStatList: ", netStatList)

			// range through the net stats
			for _, val := range netStatList {

				// calc latency with a reliable host
				go calcLatency(reliable_host, latencyChan)

				// wait for latency result and define network metrics
				networkMetrics := NetworkMetrics{
					Timestamp:       time.Now().Unix(),
					Device:          val.Name,
					Latency:         <-latencyChan,
					BytesSent:       int64(val.BytesSent),
					BytesReceived:   int64(val.BytesRecv),
					PacketsSent:     int64(val.PacketsSent),
					PacketsReceived: int64(val.PacketsRecv),
				}

				fmt.Println("Net stat val: ", val)

				// produce the msg to Kafka
				produceMsgKafka(producer, networkMetrics)

			}
		}

		time.Sleep(DELAY_SEC * time.Second)
		// --------------------------------------------------------------------------------------
	}
}

func produceMsgKafka(producer sarama.AsyncProducer, networkMetrics NetworkMetrics) {
	// serialize metrics to byte slice
	metricsBSlice, err := json.Marshal(networkMetrics)
	if err != nil {
		log.Panic("Error while converting metrics to byte slice: ", err)
	}

	// kafka message
	msg := createKafkaMessage(metricsBSlice)

	select {

	case producer.Input() <- msg:
		log.Println("==== Sent msg to kafka successfully =====", networkMetrics)
		log.Println("\n\n\n")
		handleProducerClose(producer)

		// handle errors
	case err := <-producer.Errors():
		log.Println("Error in producer: ", err)

	}
}

func createKafkaMessage(bSlice []byte) *sarama.ProducerMessage {
	bEnc := sarama.ByteEncoder(bSlice)
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   nil,
		Value: bEnc,
	}
	return msg
}

func calcLatency(reliableHost string, latencyChan chan float64) {
	// run ping to a reliable host
	pingResult, err := exec.Command("ping", "-c", "4", reliableHost).Output()
	if err != nil {
		log.Panicln("Error while pinging host:", err)
	}

	// convert result to string
	pingStr := string(pingResult)
	var match []string

	log.Println("pingStr: ", pingStr)

	regExp1 := `round-trip min/avg/max/stddev = [\d\.]+/([\d\.]+)/[\d\.]+/[\d\.]+ ms`
	regExp2 := `round-trip min\/avg\/max = [\d\.]+\/([\d\.]+)\/[\d\.]+ ms`

	if isMatching(regExp1, pingStr, match, latencyChan) ||
		isMatching(regExp2, pingStr, match, latencyChan) {
		log.Println("latency check done...")
	} else {
		// default value
		log.Println("Failed to calculate latency, using default value 0")
		latencyChan <- 0
	}
}

func isMatching(regExp, pingStr string, match []string, latencyChan chan float64) bool {
	result := false
	rExp := regexp.MustCompile(regExp)
	log.Println("rExp:", rExp)
	match = rExp.FindStringSubmatch(pingStr)
	log.Println("match:", match)
	if len(match) < 2 {
		log.Println("Failed to parse ping output: ", pingStr)
		result = false
	} else {
		result = true
		// Convert the average latency to float64
		avgLatency, err := strconv.ParseFloat(match[1], 64)
		if err != nil {
			log.Panicln("Failed to convert latency to float: v", err)
		}
		log.Println("avgLatency", avgLatency)
		latencyChan <- avgLatency
	}
	return result
}
