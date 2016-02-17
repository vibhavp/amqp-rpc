package amqprpc_test

import (
	"log"
	"net/rpc"

	"github.com/streadway/amqp"
	"github.com/vibhavp/amqp-rpc"
)

func ExampleNewServerCodec() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal(err)
	}

	serverCodec, err := amqprpc.NewServerCodec(conn, "rpc_queue", amqprpc.GobCodec{})
	if err != nil {
		log.Fatal(err)
	}

	go rpc.ServeCodec(serverCodec)
}

func ExampleNewClientCodec() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal(err)
	}

	clientCodec, err := amqprpc.NewClientCodec(conn, "rpc_queue", amqprpc.GobCodec{})
	if err != nil {
		log.Fatal(err)
	}

	client := rpc.NewClientWithCodec(clientCodec)
	client.Call("Foo.Bar", struct{}{}, &struct{}{})
}
