package nevent

import (
	"context"
	"fmt"
	"sync"

	pb "github.com/LilithGames/nevent/proto"
	"github.com/golang/protobuf/proto"
	"github.com/nats-io/nats.go"
)

type clientOptions struct {
	interceptor        ClientInterceptor
	subjectTransformer SubjectTransformer
}

type ClientOption interface {
	apply(*clientOptions)
}

type funcClientOption struct {
	f func(*clientOptions)
}

func (it *funcClientOption) apply(o *clientOptions) {
	it.f(o)
}

func newFuncClientOption(f func(*clientOptions)) *funcClientOption {
	return &funcClientOption{f: f}
}

func ClientSubjectTransformer(ts SubjectTransformer) ClientOption {
	return newFuncClientOption(func(o *clientOptions) {
		o.subjectTransformer = ts
	})
}

func ClientSTValue(value string) ClientOption {
	return newFuncClientOption(func(o *clientOptions) {
		o.subjectTransformer = STValue(value)
	})
}

type emitOptions struct {
	subjectTransformer SubjectTransformer
}

type EmitOption interface {
	apply(*emitOptions)
}

type funcEmitOption struct {
	f func(*emitOptions)
}

func (it *funcEmitOption) apply(o *emitOptions) {
	it.f(o)
}

func newFuncEmitOption(f func(*emitOptions)) *funcEmitOption {
	return &funcEmitOption{f: f}
}

func EmitSubjectTransformer(ts SubjectTransformer) EmitOption {
	return newFuncEmitOption(func(o *emitOptions) {
		o.subjectTransformer = ts
	})
}

func EmitSTValue(value string) EmitOption {
	return newFuncEmitOption(func(o *emitOptions) {
		o.subjectTransformer = STValue(value)
	})
}

type Client struct {
	nc   *nats.Conn
	o    *clientOptions
	once *sync.Once
	jet  nats.JetStreamContext
	jerr error
}

func NewClient(nc *nats.Conn, opts ...ClientOption) (*Client, error) {
	o := &clientOptions{}
	for _, opt := range opts {
		opt.apply(o)
	}
	if o.interceptor == nil {
		o.interceptor = IdentityClientInterceptor()
	}
	if o.subjectTransformer == nil {
		o.subjectTransformer = DefaultSubjectTransformer()
	}
	return &Client{
		nc:   nc,
		o:    o,
		once: &sync.Once{},
	}, nil
}

func (it *Client) Jet() (nats.JetStreamContext, error) {
	it.once.Do(func() {
		it.jet, it.jerr = it.nc.JetStream()
	})
	return it.jet, it.jerr
}

func (it *Client) GetOptions(opts ...EmitOption) *emitOptions {
	o := &emitOptions{}
	for _, opt := range opts {
		opt.apply(o)
	}
	if o.subjectTransformer == nil {
		o.subjectTransformer = it.o.subjectTransformer
	}
	return o
}

func (it *Client) GetSubject(subject string, o *emitOptions) string {
	return o.subjectTransformer(subject)
}

func (it *Client) Emit(ctx context.Context, m *nats.Msg, opts ...EmitOption) error {
	o := it.GetOptions(opts...)
	m.Subject = it.GetSubject(m.Subject, o)
	next := func(ctx context.Context, t pb.EventType, m *nats.Msg) (interface{}, error) {
		err := it.nc.PublishMsg(m)
		return nil, fmt.Errorf("nevent emit %s push msg %w", m.Subject, err)
	}
	_, err := it.o.interceptor(next)(ctx, pb.EventType_Event, m)
	return fmt.Errorf("nevent emit %s next proc %w", m.Subject, err)
}

func (it *Client) Ask(ctx context.Context, m *nats.Msg, opts ...EmitOption) ([]byte, error) {
	o := it.GetOptions(opts...)
	m.Subject = it.GetSubject(m.Subject, o)
	next := func(ctx context.Context, t pb.EventType, m *nats.Msg) (interface{}, error) {
		resp, err := it.nc.RequestMsgWithContext(ctx, m)
		if err != nil {
			return nil, fmt.Errorf("nevent ask request %s msg err %w", m.Subject, err)
		}
		answer := new(pb.Answer)
		err = proto.Unmarshal(resp.Data, answer)
		if err != nil {
			return nil, fmt.Errorf("nevent ask %s rsp unmarshal error: %w", m.Subject, err)
		}
		if answer.Error != "" {
			return nil, fmt.Errorf("nevent ask %s answer error: %s", m.Subject, answer.Error)
		}
		return answer.Data, nil
	}
	resp, err := it.o.interceptor(next)(ctx, pb.EventType_Ask, m)
	if err != nil {
		return nil, fmt.Errorf("nevent ask %s rsp proc err:%w", m.Subject, err)
	}
	return resp.([]byte), nil
}

func (it *Client) Push(ctx context.Context, m *nats.Msg, opts ...EmitOption) (*pb.PushAck, error) {
	o := it.GetOptions(opts...)
	m.Subject = it.GetSubject(m.Subject, o)
	next := func(ctx context.Context, t pb.EventType, m *nats.Msg) (interface{}, error) {
		jet, err := it.Jet()
		if err != nil {
			return nil, fmt.Errorf("nevent push %s jet construct %w", m.Subject, err)
		}
		_, err = jet.PublishMsg(m)
		if err != nil {
			return nil, fmt.Errorf("nevent push %s jet push msg %w", m.Subject, err)
		}
		return new(pb.PushAck), nil
	}
	resp, err := it.o.interceptor(next)(ctx, pb.EventType_Push, m)
	if err != nil {
		return nil, fmt.Errorf("nevent push %s next proc %w", m.Subject, err)
	}
	return resp.(*pb.PushAck), nil
}
