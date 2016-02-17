package amqprpc

import (
	"flag"
	"log"
	"math/rand"
	"net/rpc"
	"sync"
	"testing"
	"time"
)

var (
	url      = flag.String("url", "amqp://guest:guest@localhost:5672/", "AMQP queue URL")
	queue    = flag.String("queue", "rpc_queue", "queue name")
	rpcCalls = flag.Int("calls", 10, "number of RPC calls to do")
)

func init() {
	log.SetFlags(log.Lshortfile)
	flag.Parse()
}

type RPC int

func (RPC) Add(args Args, reply *int) error {
	r := rand.Intn(10)
	time.Sleep(time.Duration(r) * time.Millisecond)
	*reply = args.A + args.B
	return nil
}

type Args struct {
	A, B int
}

func TestRPC(b *testing.T) {
	serverCodec, err := NewServerCodec(*url, *queue, JSONCodec{})
	if err != nil {
		b.Fatal(err)
	}

	server := rpc.NewServer()

	err = server.Register(new(RPC))
	if err != nil {
		b.Fatal(err)
	}

	go func() { server.ServeCodec(serverCodec) }()

	var clientCodecs []rpc.ClientCodec
	var clients []*rpc.Client
	wait := new(sync.WaitGroup)
	mu := new(sync.Mutex)

	wait.Add(10)

	for i := 0; i < 10; i++ {
		go func() {
			codec, err := NewClientCodec(*url, *queue, JSONCodec{})
			if err != nil {
				b.Fatal(err)
			}

			mu.Lock()
			clientCodecs = append(clientCodecs, codec)
			clients = append(clients, rpc.NewClientWithCodec(codec))
			mu.Unlock()
			wait.Done()
		}()
	}

	wait.Wait()

	for i := 0; i < 10; i++ {
		wait.Add(10)
		go func() {
			for _, client := range clients {
				go doCall(b, client, wait)
			}
		}()
	}

	wait.Wait()
}

func doCall(b *testing.T, client *rpc.Client, wait *sync.WaitGroup) {
	num1, num2 := rand.Intn(10000000), rand.Intn(10000000)

	reply := new(int)
	err := client.Call("RPC.Add", Args{num1, num2}, reply)
	if err != nil {
		b.Fatal(err)
	}

	if num1+num2 != *reply {
		b.Fatalf("%d + %d != %d", num1, num2, *reply)
	}

	wait.Done()
}
