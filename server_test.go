package nevent_test

import (
	"fmt"
	"testing"
	"context"
	"time"
	"net/http"
	_ "net/http/pprof"

	"github.com/LilithGames/nevent"
	npb "github.com/LilithGames/nevent/proto"
	pb "github.com/LilithGames/nevent/testdata/proto"
	"github.com/nats-io/nats.go"

	"github.com/stretchr/testify/assert"
)

type service struct{
	id string
	queue string
}

func (it *service) OnPersonEvent(ctx context.Context, e *pb.Person) {
	fmt.Printf("event(%s, %s): %+v\n", it.id, it.queue, e)
}

func (it *service) OnPersonAsk(ctx context.Context, e *pb.Person) (*pb.Company, error) {
	fmt.Printf("ask(%s, %s): %+v\n", it.id, it.queue, e)
	return &pb.Company{Name: e.Name}, nil
}

func (it *service) OnPersonPush(ctx context.Context, e *pb.Person) error {
	fmt.Printf("push(%s, %s): %+v\n", it.id, it.queue, e)
	return nil
}

func (it *service) OnError(err error) {
	fmt.Printf("err: %+v\n", err)
}

func provideServer(t *testing.T, id string, queue string) *nevent.Server {
	s, err := nats.Connect(nats.DefaultURL)
	assert.Nil(t, err)
	i1 := func(next nevent.ServerEventHandler) nevent.ServerEventHandler {
		return func(ctx context.Context, t npb.EventType, m *nats.Msg) (interface{}, error) {
			// fmt.Printf("server i1: %+v\n", t)
			return next(ctx, t, m)
		}
	}
	i2 := func(next nevent.ServerEventHandler) nevent.ServerEventHandler {
		return func(ctx context.Context, t npb.EventType, m *nats.Msg) (interface{}, error) {
			// fmt.Printf("server i2: %+v\n", t)
			return next(ctx, t, m)
		}
	}
	interceptor := nevent.FuncServerInterceptor(nevent.ChainServerInterceptor(i1, i2))
	eh := func(err error) {
		fmt.Printf("%+v\n", err)
	}
	es, err := nevent.NewServer(s, nevent.Queue(queue), nevent.ServerErrorHandler(eh), interceptor)
	assert.Nil(t, err)
	svc := &service{id: id, queue: queue}
	pb.RegisterPersonEvent(es, pb.PersonEventFuncListener(func(ctx context.Context, m *pb.Person){
		fmt.Printf("func event(%s, %s): %+v\n", id, queue, m)

	}), nevent.ListenSTValue("*"))
	pb.RegisterPersonAsk(es, svc)
	pb.RegisterPersonPush(es, svc)
	return es
}

func provideStream(t *testing.T) {
	nc, err := nats.Connect(nats.DefaultURL)
	str, err := nevent.NewStream(nc)
	assert.Nil(t, err)
	_, err = pb.EnsurePersonPushStream(str, nevent.StreamForceUpdate())
	assert.Nil(t, err)
}

func provideClient(t *testing.T) *pb.TestClient {
	c, err := nats.Connect(nats.DefaultURL)
	assert.Nil(t, err)
	i1 := func(next nevent.ClientEventInvoker) nevent.ClientEventInvoker {
		return func(ctx context.Context, t npb.EventType, m *nats.Msg) (interface{}, error) {
			// fmt.Printf("client i1: %+v\n", t)
			return next(ctx, t, m)
		}
	}
	i2 := func(next nevent.ClientEventInvoker) nevent.ClientEventInvoker {
		return func(ctx context.Context, t npb.EventType, m *nats.Msg) (interface{}, error) {
			// fmt.Printf("client i2: %+v\n", t)
			return next(ctx, t, m)
		}
	}
	interceptor := nevent.FuncClientInterceptor(nevent.ChainClientInterceptor(i1, i2))
	nc, err := nevent.NewClient(c, interceptor)
	assert.Nil(t, err)
	return pb.NewTestClient(nc)
}

func TestEvent(t *testing.T) {
	provideServer(t, "id", "queue")
	pbc := provideClient(t)
	err := pbc.PersonEvent(context.TODO(), &pb.Person{Name: "hulucc_event"}, nevent.EmitSTValue("1"))
	assert.Nil(t, err)
	select{}
}

func TestEventLeak(t *testing.T) {
	go http.ListenAndServe("0.0.0.0:6060", nil)
	provideServer(t, "id", "queue")
	pbc := provideClient(t)
	for i := 0; ; i++ {
		pbc.PersonEvent(context.TODO(), &pb.Person{Name: fmt.Sprintf("hulucc_event%d", i)}, nevent.EmitSTValue("1"))
		time.Sleep(time.Millisecond)
	}
}


func TestAsk(t *testing.T) {
	provideServer(t, "id", "queue")
	pbc := provideClient(t)
	company, err := pbc.PersonAsk(context.TODO(), &pb.Person{Name: "hulucc_ask"})
	assert.Nil(t, err)
	fmt.Printf("answer: %+v\n", company)
	select{}
}

func TestPush(t *testing.T) {
	provideStream(t)
	provideServer(t, "id1", "queue_1")
	provideServer(t, "id2", "queue_1")
	provideServer(t, "id1", "queue_2")
	provideServer(t, "id2", "queue_2")
	pbc := provideClient(t)
	_, err := pbc.PersonPush(context.TODO(), &pb.Person{Name: "hulucc_push"})
	assert.Nil(t, err)
	select{}
}

