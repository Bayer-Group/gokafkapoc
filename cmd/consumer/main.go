package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/linkedin/goavro"
	"github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"
	"github.com/MonsantoCo/gokafkapoc/internal/avro"
	"github.com/MonsantoCo/gokafkapoc/internal/stream"
	"github.com/MonsantoCo/gokafkapoc/internal/zipcode"
)

const (
	host              = "192.168.99.100"
	baseurl           = "http://" + host
	zipTopic          = "zip"
	ziprichTopic      = "zip_address"
	maxNumberRequests = 150
)

var (
	zipCodec     *goavro.Codec
	ziprichCodec *goavro.Codec
)

func init() {
	client := avro.MustGetClient(baseurl + ":8081")
	zipCodec = client.MustGetLatestCodec("zipcode-value")
	ziprichCodec = client.MustGetLatestCodec("zip_address-value")
}

func main() {
	ctx := NewContextWithShutdown()
	producerStream := make(chan kafka.Message)

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{host + ":9092"},
		GroupID: "zipRichers",
		Topic:   zipTopic,
	})
	go func() {
		semaphore := make(chan struct{}, maxNumberRequests)
		defer close(semaphore)

		for {
			m, err := r.ReadMessage(ctx)
			if err != nil {
				log.Error(err)
				return
			}

			var zip zipcode.Zipcode
			if err := avro.Decode(m.Value, &zip, zipCodec); err != nil {
				log.Error(err)
				continue
			}

			semaphore <- struct{}{} // adds request to semaphore, if semaphore chan buffer is full, will block here
			go func() {
				defer func() { <-semaphore }() // removes request from semaphore making space for next request
				if err := stream.Enrich(ctx, producerStream, zip, "http://127.0.0.1:3113/zipcode", ziprichCodec); err != nil {
					log.Error(err)
				}
			}()
		}
	}()

	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{host + ":9092"},
		Topic:    ziprichTopic,
		Balancer: &kafka.RoundRobin{},
	})
	defer w.Close()
	metrics := stream.ProcessMessages(ctx, w, producerStream, ziprichTopic)

	fmt.Printf("\t%s >\tsuccesses: %d\terrors: %d\ttotal processed msgs: %d\n", ziprichTopic, metrics.Successes, metrics.Errs, metrics.MsgCount)
}

func NewContextWithShutdown() context.Context {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-shutdown
		cancel()
	}()

	return ctx
}
