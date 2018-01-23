package zipcode

import (
	"context"
	"time"

	"github.com/icrowley/fake"
	"github.com/linkedin/goavro"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"
)

type Ziprich struct {
	Zipcode
	Addresses
}

func (zr Ziprich) Map() map[string]interface{} {
	out := zr.Zipcode.Map()
	out["addresses"] = zr.Addresses.Collect()
	return out
}

type Addresses []Address

func (adds Addresses) Collect() []interface{} {
	out := make([]interface{}, 0)
	for _, address := range adds {
		out = append(out, address.Map())
	}
	return out
}

type Address struct {
	Zipcode string
	Street  string
	City    string
	State   string
	Country string
}

func (a Address) Map() map[string]interface{} {
	out := make(map[string]interface{})
	out["zipcode"] = a.Zipcode
	out["street"] = a.Street
	out["city"] = a.City
	out["state"] = a.State
	out["country"] = a.Country
	return out
}

func ZiprichGenerator(ctx context.Context, messages chan<- kafka.Message, valCodec *goavro.Codec) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			zipRich := Ziprich{
				Zipcode: Zipcode{fake.Zip(), time.Now().UnixNano()},
				Addresses: Addresses{
					{"92647", "Lindbergh Drive", "Creve Coeur", "MO", "US"},
					{"63108", "Lindell Blvd", "Saint Louis", "MO", "US"},
					{"63103", "Lindell Blvd", "Saint Louis", "MO", "US"},
				},
			}
			ziprichOut, err := valCodec.TextualFromNative(nil, zipRich.Map())
			if err != nil {
				log.Error(errors.Wrap(err, "error in ziprich codec"))
				continue
			}
			messages <- kafka.Message{Value: ziprichOut}
		}
	}
}
