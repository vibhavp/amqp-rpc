amqp-rpc
========

[![Go Report Card](https://goreportcard.com/badge/github.com/vibhavp/amqp-rpc)](https://goreportcard.com/report/github.com/vibhavp/amqp-rpc)
[![Coverage Status](https://coveralls.io/repos/github/vibhavp/amqp-rpc/badge.svg?branch=master)](https://coveralls.io/github/vibhavp/amqp-rpc?branch=master)
[![GoDoc](https://godoc.org/github.com/vibhavp/amqp-rpc?status.svg)](https://godoc.org/github.com/vibhavp/amqp-rpc)

# Example

## Server

```go
package server

import (
	"log"
	"net/rpc"

	"github.com/streadway/amqp"
	"github.com/vibhavp/amqp-rpc"
)
	
type Args struct {
	A, B int
}

type Quotient struct {
	Quo, Rem int
}

type Arith int

func (t *Arith) Multiply(args *Args, reply *int) error {
	*reply = args.A * args.B
	return nil
}

func (t *Arith) Divide(args *Args, quo *Quotient) error {
	if args.B == 0 {
		return errors.New("divide by zero")
	}
	quo.Quo = args.A / args.B
	quo.Rem = args.A % args.B
	return nil
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal(err)
	}

	serverCodec, err := amqprpc.NewServerCodec(conn, "rpc_queue", amqprpc.GobCodec{})
	if err != nil {
		log.Fatal(err)
	}

	rpc.Register(new(Arith))
	rpc.ServeCodec(serverCodec)
}
```

## Client

```go
package client

import (
	"github.com/streadway/amqp"
	"github.com/vibhavp/amqp-rpc"
)
	
func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal(err)
	}

	clientCodec, err := amqprpc.NewClientCodec(conn, "rpc_queue", amqprpc.GobCodec{})
	if err != nil {
		log.Fatal(err)
	}

	reply := new(int)
	client := rpc.NewClientWithCodec(clientCodec)
	err := client.Call("Arith.Multiply", server.Args{1, 2}, reply)
	if err != nil {
		log.Fatal("arith error: ", err)
	}
	
	log.Printf("1 * 2 = %d", *reply)
}
```
