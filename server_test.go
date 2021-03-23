package nevent_test

import (
	"fmt"
	"testing"
	"context"

	"github.com/LilithGames/nevent"
	"google.golang.org/grpc/metadata"
	pb "github.com/LilithGames/nevent/testdata/proto"
	pbi "google.golang.org/protobuf/runtime/protoiface"
	"github.com/nats-io/nats.go"

	"github.com/stretchr/testify/assert"
)

type service struct{}

func (it *service) OnPersonEvent(ctx context.Context, e *pb.Person) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		fmt.Printf("md: %+v\n", md)
	}
	fmt.Printf("event: %+v\n", e)
}

func (it *service) OnError(err error) {
	fmt.Printf("err: %+v\n", err)
}

func provideServer() *nevent.Server {
	s, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		panic(err)
	}
	es, err := nevent.NewServer(s, nevent.ServerInterceptor(func(ctx context.Context, subject string, reply interface{}, e pbi.MessageV1) {
		fmt.Printf("interceptor subject: %+v\n", subject)
	}))
	if err != nil {
		panic(err)
	}
	svc := &service{}
	pb.RegisterPersonEvent(es, svc)
	return es
}

func provideClient() *nevent.Client {
	c, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		panic(err)
	}
	nc, err := nevent.NewClient(c)
	if err != nil {
		panic(err)
	}
	return nc
}

func TestAll(t *testing.T) {
	provideServer()
	nc := provideClient()
	ctx := metadata.NewOutgoingContext(context.TODO(), metadata.Pairs("hello", "world"))
	p := &pb.Person{Name: "hulucc"}
	err := p.Emit(ctx, nc)
	assert.Nil(t, err)
	select{}
}
