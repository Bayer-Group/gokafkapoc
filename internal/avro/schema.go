package avro

import (
	"log"

	"github.com/datamountaineer/schema-registry"
	"github.com/linkedin/goavro"
)

type Client struct {
	client schemaregistry.Client
}

func (c Client) MustGetLatestCodec(subject string) *goavro.Codec {
	schema := MustGetLatestSchema(c.client, subject)
	return MustGetCodec(schema.Schema)
}

func MustGetCodec(schema string) *goavro.Codec {
	codec, err := goavro.NewCodec(schema)
	if err != nil {
		log.Fatal(err)
	}
	return codec
}

func MustGetClient(baseurl string) Client {
	client, err := schemaregistry.NewClient(baseurl)
	if err != nil {
		log.Fatal(err)
	}
	return Client{client}
}

func MustGetLatestSchema(client schemaregistry.Client, subject string) schemaregistry.Schema {
	schema, err := client.GetLatestSchema(subject)
	if err != nil {
		log.Fatal(err)
	}
	return schema
}
