package nevent

import (
	"fmt"
	"context"
	"runtime/debug"
	"errors"
	"sync"

	"github.com/nats-io/nats.go"
	"github.com/golang/protobuf/proto"
	pb "github.com/LilithGames/nevent/proto"
)

var ErrNak error = errors.New("Nak")

type serverOptions struct {
	queue string
	interceptor ServerInterceptor
	errorHandler func(error)
	subjectTransformer SubjectTransformer
}

type ServerOption interface {
	apply(*serverOptions)
}

type funcServerOption struct {
	f func(*serverOptions)
}

func (it *funcServerOption) apply(o *serverOptions) {
	it.f(o)
}

func newFuncServerOption(f func(*serverOptions)) ServerOption {
	return &funcServerOption{f: f}
}

func Queue(queue string) ServerOption {
	return newFuncServerOption(func(o *serverOptions) {
		o.queue = queue
	})
}

func ServerSubjectTransformer(ts SubjectTransformer) ServerOption {
	return newFuncServerOption(func(o *serverOptions) {
		o.subjectTransformer = ts
	})
}

type Server struct {
	nc *nats.Conn
	o *serverOptions
	once *sync.Once
	jet nats.JetStreamContext
	jeterr error
}

// NewListener server with the same queue will listen event round-robin
func NewServer(nc *nats.Conn, opts ...ServerOption) (*Server, error) {
	o := &serverOptions{}
	for _, opt := range opts {
		opt.apply(o)
	}
	if o.errorHandler == nil {
		o.errorHandler = func(error){}
	}
	if o.interceptor == nil {
		o.interceptor = IdentityServerInterceptor()
	}
	if o.subjectTransformer == nil {
		o.subjectTransformer = DefaultSubjectTransformer()
	}
	return &Server{
		nc: nc,
		o: o,
		once: &sync.Once{},
	}, nil
}

type listenOptions struct{
	queue string
	subjectTransformer SubjectTransformer
}
type ListenOption interface {
	apply(*listenOptions)
}
type funcListenOption struct {
	f func(*listenOptions)
}
func (it *funcListenOption) apply(o *listenOptions) {
	it.f(o)
}
func newFuncListenOption(f func(*listenOptions)) ListenOption {
	return &funcListenOption{f: f}
}

func ListenQueue(queue string) ListenOption {
	return newFuncListenOption(func(o *listenOptions){
		o.queue = queue
	})
}

func ListenSubjectTransformer(ts SubjectTransformer) ListenOption {
	return newFuncListenOption(func(o *listenOptions) {
		o.subjectTransformer = ts
	})
}

func (it *Server) GetListenOptions(opts ...ListenOption) *listenOptions {
	o := &listenOptions{}
	for _, opt := range opts {
		opt.apply(o)
	}
	if o.queue == "" {
		o.queue = it.o.queue
	}
	if o.subjectTransformer == nil {
		o.subjectTransformer = it.o.subjectTransformer
	}
	return o
}

func (it *Server) GetSubject(subject string, o *listenOptions) string {
	return o.subjectTransformer(subject)
}

type EventHandler func(ctx context.Context, m *nats.Msg) (interface{}, error)

func (it *Server) ListenEvent(subject string, t pb.EventType, eh EventHandler, opts ...ListenOption) (*nats.Subscription, error) {
	o := it.GetListenOptions(opts...)
	fullsubject := it.GetSubject(subject, o)
	next := func(ctx context.Context, t pb.EventType, m *nats.Msg) (interface{}, error) {
		return eh(ctx, m)
	}

	mh := func(m *nats.Msg) {
		defer func() {
			if r := recover(); r != nil {
				it.o.errorHandler(fmt.Errorf("event handler panic: %w", errors.New(string(debug.Stack()))))
			}
		}()
		resp, err := it.o.interceptor(next)(context.TODO(), t, m)
		switch t {
		case pb.EventType_Event:
			if err != nil {
				it.o.errorHandler(err)
			}
			return
		case pb.EventType_Ask:
			answer := new(pb.Answer)
			if err != nil {
				answer.Error = err.Error()
			} else {
				answer.Data = resp.([]byte)
			}
			bs, err := proto.Marshal(answer)
			if err != nil {
				it.o.errorHandler(fmt.Errorf("server answer marshal error: %w", err))
				return
			}
			err = m.Respond(bs)
			if err != nil {
				it.o.errorHandler(fmt.Errorf("answer error: %w", err))
			}
			return
		case pb.EventType_Push:
			if errors.Is(err, ErrNak) {
				err = m.Nak()
			} else if err == nil {
				err = m.Ack()
			} else {
				it.o.errorHandler(fmt.Errorf("push error: %w", err))
				return
			}
			if err != nil {
				it.o.errorHandler(fmt.Errorf("push ack error: %w", err))
				return
			}
		default:
			panic(fmt.Errorf("not supported type: %v", t))
		}
	}
	if t == pb.EventType_Event || t == pb.EventType_Ask {
		if o.queue != "" {
			return it.nc.QueueSubscribe(fullsubject, it.o.queue, mh)
		} else {
			return it.nc.Subscribe(fullsubject, mh)
		}
	} else if t == pb.EventType_Push {
		jet, err := it.Jet()
		if err != nil {
			return nil, err
		}
		if o.queue != "" {
			return jet.QueueSubscribe(fullsubject, o.queue, mh, nats.Durable(o.queue))
		} else {
			return jet.Subscribe(fullsubject, mh)
		}
	} else {
		return nil, fmt.Errorf("unknown type: %v", t)
	}
}

func (it *Server) Jet() (nats.JetStreamContext, error) {
	it.once.Do(func() {
		it.jet, it.jeterr = it.nc.JetStream()
	})
	return it.jet, it.jeterr
}
