package amqprpc

import (
	"errors"
	"net/rpc"
	"strconv"

	"github.com/streadway/amqp"
)

type codec struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	routing string //routing key

	codec       EncodingCodec
	curDelivery amqp.Delivery

	message <-chan amqp.Delivery
}

type clientCodec struct {
	*codec
	serverRouting string //server routing key
}

func (client *clientCodec) WriteRequest(r *rpc.Request, v interface{}) error {
	//TODO(vibhavp): Set ContentType, ContentEncoding
	body, err := client.codec.codec.Marshal(v)
	if err != nil {
		return err
	}

	publishing := amqp.Publishing{
		ReplyTo:       r.ServiceMethod,
		CorrelationId: client.routing,
		MessageId:     strconv.FormatUint(r.Seq, 10),
		Body:          body,
	}

	return client.channel.Publish("", client.serverRouting, false, false, publishing)
}

func (client *clientCodec) ReadResponseHeader(r *rpc.Response) error {
	client.curDelivery = <-client.message

	var err error
	if err = client.curDelivery.Headers.Validate(); err != nil {
		return errors.New("error while reading body: " + err.Error())
	}

	if err, ok := client.curDelivery.Headers["error"]; ok {
		errStr, ok := err.(string)
		if !ok {
			return errors.New("error header not a string")
		}
		r.Error = errStr
	}

	r.ServiceMethod = client.curDelivery.ReplyTo
	r.Seq, err = strconv.ParseUint(client.curDelivery.MessageId, 10, 64)
	if err != nil {
		return err
	}

	return nil
}

func (client *clientCodec) ReadResponseBody(v interface{}) error {
	if v == nil {
		return nil
	}

	return client.codec.codec.Unmarshal(client.curDelivery.Body, v)
}

func (client *clientCodec) Close() error {
	return client.conn.Close()
}

//NewClientCodec returns a new rpc.ClientCodec using AMQP on conn. serverRouting is the routing
//key with with RPC calls are sent, it should be the same routing key used with NewServerCodec.
//encodingCodec is an EncodingCoding implementation. This package provdes JSONCodec and GobCodec
//for the JSON and Gob encodings respectively.
func NewClientCodec(conn *amqp.Connection, serverRouting string, encodingCodec EncodingCodec) (rpc.ClientCodec, error) {
	channel, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	queue, err := channel.QueueDeclare("", false, true, false, false, nil)
	if err != nil {
		return nil, err
	}

	message, err := channel.Consume(queue.Name, "", true, false, false, false, nil)
	client := &clientCodec{
		codec: &codec{
			conn:    conn,
			channel: channel,
			routing: queue.Name,
			codec:   encodingCodec,
			message: message,
		},

		serverRouting: serverRouting,
	}

	return client, err
}
