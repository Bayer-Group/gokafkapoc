package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"

	"github.com/linkedin/goavro"
	"github.com/segmentio/kafka-go"

	"github.com/MonsantoCo/gokafkapoc/internal/avro"
	"github.com/MonsantoCo/gokafkapoc/internal/monitor"
	"github.com/MonsantoCo/gokafkapoc/internal/stream"
	"github.com/MonsantoCo/gokafkapoc/internal/zipcode"
)

const (
	host     = "192.168.99.100"
	baseurl  = "http://" + host
	zipTopic = "zip"
)

var zipCodec *goavro.Codec

func init() {
	client := avro.MustGetClient(baseurl + ":8081")
	zipCodec = client.MustGetLatestCodec("zipcode-value")
}

func main() {
	ctx := NewContextWithShutdown()
	numWorkers := 3

	topicMetrics := make(monitor.MetricsByTopic)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		w := kafka.NewWriter(kafka.WriterConfig{
			Brokers:  []string{host + ":9092"},
			Topic:    zipTopic,
			Balancer: &kafka.RoundRobin{},
		})
		defer w.Close()

		zipMessageStream := NewMessageStream(ctx, numWorkers, zipcode.ZipcodeGenerator, zipCodec)
		metrics := stream.ProcessMessages(ctx, w, zipMessageStream, zipTopic)
		topicMetrics[zipTopic] = metrics
	}()

	wg.Wait()
	fmt.Println(topicMetrics)
}

type Generator func(ctx context.Context, messages chan<- kafka.Message, codec *goavro.Codec)

func NewMessageStream(ctx context.Context, numWorkers int, gen Generator, codec *goavro.Codec) <-chan kafka.Message {
	messageStream := make(chan kafka.Message)

	go func() {
		defer close(messageStream)
		var wg sync.WaitGroup
		wg.Add(numWorkers)
		for i := 0; i < numWorkers; i++ {
			go func() {
				defer wg.Done()
				gen(ctx, messageStream, codec)
			}()
		}
		wg.Wait()
	}()

	return messageStream
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
