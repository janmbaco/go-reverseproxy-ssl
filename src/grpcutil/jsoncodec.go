package grpcutil

import (
	"bytes"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"google.golang.org/grpc/encoding"
)

func init() {
	encoding.RegisterCodec(JSON{
		Marshaler: jsonpb.Marshaler{
			EmitDefaults: true,
			OrigName:     true,
		},
	})
}

// JSON structure is used to marshall or unmarshal a protobuffer messsage.
type JSON struct {
	jsonpb.Marshaler
	jsonpb.Unmarshaler
}

// Name returns "json"
func (JSON) Name() string {
	return "json"
}

// Marshal returns proto message from json format.
func (json JSON) Marshal(v interface{}) (out []byte, err error) {
	if pm, ok := v.(proto.Message); ok {
		b := new(bytes.Buffer)
		err := json.Marshaler.Marshal(b, pm)
		if err != nil {
			return nil, err
		}
		return b.Bytes(), nil
	}
	return json.Marshal(v)
}

// Unmarshal returns a json message from protobuffer.
func (json JSON) Unmarshal(data []byte, v interface{}) (err error) {
	if pm, ok := v.(proto.Message); ok {
		b := bytes.NewBuffer(data)
		return json.Unmarshaler.Unmarshal(b, pm)
	}
	return json.Unmarshal(data, v)
}