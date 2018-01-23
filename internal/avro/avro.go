package avro

import (
	"github.com/linkedin/goavro"
	"github.com/pkg/errors"
)

type Mapper interface {
	Map() map[string]interface{}
}

type Unmapper interface {
	Unmap(interface{}) error
}

func Encode(v Mapper, codec *goavro.Codec) (out []byte, err error) {
	out, err = codec.TextualFromNative(nil, v.Map())
	err = errors.Wrap(err, "error in zip codec")
	return
}

func Decode(b []byte, dest Unmapper, codec *goavro.Codec) error {
	native, _, err := codec.NativeFromTextual(b)
	if err != nil {
		return err
	}
	return dest.Unmap(native)
}
