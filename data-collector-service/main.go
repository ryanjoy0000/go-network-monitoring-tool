package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strconv"
	"time"

	"github.com/IBM/sarama"
	"github.com/joho/godotenv"
	"github.com/shirou/gopsutil/v4/net"
)

const (
	ENV = ".env"
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
	// load .env file
	loadEnv()

	// fetch vars from .env
	topic = os.Getenv("KAFKA_TOPIC")
	addrs = []string{os.Getenv("KAFKA_HOST")}
	reliable_host = os.Getenv("RELIABLE_HOST")

	log.Println(".env: \n", addrs, "\n", topic, "\n", reliable_host)

	// create kafka producer
	producer, err := sarama.NewAsyncProducer(addrs, nil)
	if err != nil {
		log.Panicln("Error while creating producer", err)
	}
	log.Println("kafka producer ready...")

	// close producer properly in the end
	defer handleProducerClose(producer)

	// start collection of network metrics
	startDataCollection(producer)
}

func loadEnv() {
	err := godotenv.Load(ENV)
	if err != nil {
		log.Fatalln("Error while reading .env file", err)
	}
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
		// get network I/O statistics for every network interface installed on the system
		netStatList, err := net.IOCounters(false)
		if err != nil {
			log.Panicln("Error while collecting network metrics: ", err)
			canRun = false
		} else {
			// range through the net stats
			for _, val := range netStatList {

				// calc latency with a reliable host
				go calcLatency(reliable_host, latencyChan)

				// wait for latency result and define network metrics
				metrics := NetworkMetrics{
					Timestamp: time.Now().Unix(),
					Device:    val.Name,
					Latency:   <-latencyChan,
					Bandwidth: float64(val.BytesRecv+val.BytesSent) / 1024, // KBps
				}

				fmt.Println("metrics: ", metrics)

				// serialize metrics to byte slice
				metricsBSlice, err := json.Marshal(metrics)
				if err != nil {
					log.Panic("Error while converting metrics to byte slice: ", err)
				}

				// kafka message
				msg := createKafkaMessage(metricsBSlice)

				// produce the msg to Kafka
				produceMsgKafka(msg, producer)

			}
		}
	}
}

func produceMsgKafka(msg *sarama.ProducerMessage, producer sarama.AsyncProducer) {
	// Trap SIGINT to trigger a shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	var enqueuedCount, producerErrCount int

ProducerLoop:
	for {
		select {
		case producer.Input() <- msg:
			enqueuedCount++
		case err := <-producer.Errors():
			log.Println("Could not produce msg to kafka: ", err)
			producerErrCount++
		case <-signalChan:
			log.Println("Interruption.. Producer exiting...")
			break ProducerLoop
		}
	}

	log.Println("Total msgs enqueued:", enqueuedCount)
	log.Println("Total producer errors:", producerErrCount)
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
