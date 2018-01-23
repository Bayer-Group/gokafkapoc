package zipcode

import (
	"context"
	"net/http"
	"time"

	"bytes"
	"encoding/json"
	"io"

	"github.com/icrowley/fake"
	"github.com/linkedin/goavro"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"
	"github.com/MonsantoCo/gokafkapoc/internal/avro"
)

type Zipcode struct {
	Zipcode   string
	Timestamp int64
}

func NewRandomZipcode() Zipcode {
	return Zipcode{fake.Zip(), time.Now().UnixNano()}
}

func (z Zipcode) Map() map[string]interface{} {
	output := make(map[string]interface{})
	output["zipcode"] = z.Zipcode
	output["timestamp"] = z.Timestamp
	return output
}

func (z *Zipcode) Unmap(data interface{}) error {
	mappedZip, ok := data.(map[string]interface{})
	if !ok {
		return errors.Errorf("failed to unmap map zipcode, got %v", mappedZip)
	}
	z.Zipcode = mappedZip["zipcode"].(string)
	z.Timestamp = mappedZip["timestamp"].(int64)
	return nil
}

func (z Zipcode) Enrich(ctx context.Context, endpoint string) (avro.Mapper, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(z); err != nil {
		return Ziprich{}, errors.Wrap(err, "error marshaling zipcode")
	}

	resp, err := postEnricher(ctx, endpoint, &buf)
	if err != nil {
		return Ziprich{}, errors.Wrap(err, "error in ziprich request")
	}
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)

	var zipriched Ziprich
	err = dec.Decode(&zipriched)

	return zipriched, errors.Wrap(err, "unable to decode into ziprich type")
}

func postEnricher(ctx context.Context, endpoint string, b io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("POST", endpoint, b)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := http.DefaultClient
	return client.Do(req.WithContext(ctx))
}

func ZipcodeGenerator(ctx context.Context, messages chan<- kafka.Message, valCodec *goavro.Codec) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			zipcode := NewRandomZipcode()
			zipOut, err := valCodec.TextualFromNative(nil, zipcode.Map())
			if err != nil {
				log.Println(errors.Wrap(err, "error in zip codec"))
				continue
			}
			messages <- kafka.Message{Value: zipOut}
		}
	}
}
