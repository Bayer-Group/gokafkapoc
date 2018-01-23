package stream

import (
	"context"

	"github.com/linkedin/goavro"
	"github.com/segmentio/kafka-go"
	"github.com/MonsantoCo/gokafkapoc/internal/avro"
)

type Enricher interface {
	Enrich(ctx context.Context, endpoint string) (avro.Mapper, error)
}

func Enrich(ctx context.Context, stream chan<- kafka.Message, e Enricher, endpoint string, codec *goavro.Codec) error {
	enriched, err := e.Enrich(ctx, endpoint)
	if err != nil {
		return err
	}
	val, err := avro.Encode(enriched, codec)
	if err != nil {
		return err
	}
	stream <- kafka.Message{Value: val}
	return nil
}
