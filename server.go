package amqprpc

import (
	"errors"
	"net/rpc"
	"sync"
	"sync/atomic"

	"github.com/streadway/amqp"
)

type route struct {
	messageID string
	routing   string
}

type serverCodec struct {
	*codec
	lock  *sync.RWMutex
	calls map[uint64]route
	seq   uint64
}

func (server *serverCodec) ReadRequestHeader(r *rpc.Request) error {
	server.curDelivery = <-server.message

	if server.curDelivery.CorrelationId == "" {
		return errors.New("no routing key in delivery")
	}

	r.Seq = atomic.AddUint64(&server.seq, 1)

	server.lock.Lock()
	server.calls[r.Seq] = route{server.curDelivery.MessageId, server.curDelivery.CorrelationId}
	server.lock.Unlock()

	r.ServiceMethod = server.curDelivery.ReplyTo

	return nil
}

func (server *serverCodec) ReadRequestBody(v interface{}) error {
	if v == nil {
		return nil
	}

	return server.codec.codec.Unmarshal(server.curDelivery.Body, v)
}

func (server *serverCodec) WriteResponse(resp *rpc.Response, v interface{}) error {
	body, err := server.codec.codec.Marshal(v)
	if err != nil {
		return err
	}

	server.lock.RLock()
	route, ok := server.calls[resp.Seq]
	server.lock.RUnlock()
	if !ok {
		return errors.New("sequence dosen't have a route")
	}

	publishing := amqp.Publishing{
		ReplyTo:       resp.ServiceMethod,
		MessageId:     route.messageID,
		CorrelationId: route.routing,
		Body:          body,
	}

	if resp.Error != "" {
		publishing.Headers = amqp.Table{"error": resp.Error}
	}

	return server.channel.Publish(
		"",
		route.routing,
		false,
		false,
		publishing,
	)
}
func (server *serverCodec) Close() error {
	return server.conn.Close()
}

//NewServerCodec returns a new rpc.ClientCodec using AMQP on conn. serverRouting is the routing
//key with with RPC calls are received, encodingCodec is an EncodingCoding implementation. This package provdes JSONCodec and GobCodec for the JSON and Gob encodings respectively.
func NewServerCodec(conn *amqp.Connection, serverRouting string, encodingCodec EncodingCodec) (rpc.ServerCodec, error) {
	channel, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	queue, err := channel.QueueDeclare(serverRouting, true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	messages, err := channel.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	server := &serverCodec{
		codec: &codec{
			conn:    conn,
			channel: channel,
			routing: queue.Name,

			codec:   encodingCodec,
			message: messages,
		},
		lock:  new(sync.RWMutex),
		calls: make(map[uint64]route),
		seq:   0,
	}

	return server, err
}
