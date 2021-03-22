package nevent_test

import (
	"fmt"
	"testing"
	"context"

	"github.com/LilithGames/nevent"
	pb "github.com/LilithGames/nevent/testdata/proto"
	"github.com/nats-io/nats.go"
)

type service struct{}

func (it *service) OnPersonEvent(ctx context.Context, e *pb.Person) {
	fmt.Printf("%+v\n", e)
}

func provideServer() *nevent.Server {
	s, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		panic(err)
	}
	es, err := nevent.NewServer(s)
	if err != nil {
		panic(err)
	}
	svc := &service{}
	pb.ListenPersonEvent(es, svc)
	return es
}

func provideClient() *pb.Client {
	c, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		panic(err)
	}
	ec, err := pb.NewClient(c)
	if err != nil {
		panic(err)
	}
	return ec
}

func TestAll(t *testing.T) {
	s := provideServer()
	c := provideClient()
	println(s, c)
}
