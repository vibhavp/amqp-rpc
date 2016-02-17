package amqprpc

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
)

//EncodingCodec implements marshaling and unmarshaling of seralized data.
type EncodingCodec interface {
	Marshal(interface{}) ([]byte, error)
	Unmarshal([]byte, interface{}) error
}

//JSONCodec is an EncodingCodec implementation to send/receieve JSON data
//over AMQP.
type JSONCodec struct{}

func (JSONCodec) Marshal(v interface{}) ([]byte, error) {
	b, err := json.Marshal(v)
	return b, err
}

func (JSONCodec) Unmarshal(data []byte, v interface{}) error {
	err := json.Unmarshal(data, v)
	return err
}

//GobCodec is an EncodingCodec implementation to send/recieve Gob data
//over AMQP.
type GobCodec struct{}

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
