package amqprpc

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
)

type EncodingCodec interface {
	Marshal(interface{}) ([]byte, error)
	Unmarshal([]byte, interface{}) error
}

type JSONCodec struct{}
type GobCodec struct{}

func (JSONCodec) Marshal(v interface{}) ([]byte, error) {
	b, err := json.Marshal(v)
	return b, err
}

func (JSONCodec) Unmarshal(data []byte, v interface{}) error {
	err := json.Unmarshal(data, v)
	return err
}

func (GobCodec) Marshal(v interface{}) ([]byte, error) {
	body := new(bytes.Buffer)
	enc := gob.NewEncoder(body)
	err := enc.Encode(v)
	return body.Bytes(), err
}

func (GobCodec) Unmarshal(data []byte, v interface{}) error {
	body := bytes.NewBuffer(data)
	dec := gob.NewDecoder(body)
	err := dec.Decode(v)
	return err
}
