package nevent

import (
	"github.com/nats-io/nats.go"
	pb "github.com/LilithGames/nevent/proto"
	natspb "github.com/nats-io/nats.go/encoders/protobuf"
)

type clientOptions struct{
	interceptor ClientEventInterceptor
}

type ClientOption interface{
	apply(*clientOptions)
}

type funcClientOption struct{
	f func(*clientOptions)
}

func (it *funcClientOption) apply(o *clientOptions) {
	it.apply(o)
}

func newFuncClientOption(f func(*clientOptions)) *funcClientOption {
	return &funcClientOption{f: f}
}

type emitOptions struct{}

type EmitOption interface {
	apply(*emitOptions)
}

type funcEmitOption struct{
	f func(*emitOptions)
}

func (it *funcEmitOption) apply(o *emitOptions) {
	it.apply(o)
}

func newFuncEmitOption(f func(*emitOptions)) *funcEmitOption {
	return &funcEmitOption{f: f}
}

type Client struct {
	ec *nats.EncodedConn
	o *clientOptions
}

func NewClient(nc *nats.Conn, opts ...ClientOption) (*Client, error) {
	ec, err := nats.NewEncodedConn(nc, natspb.PROTOBUF_ENCODER)
	if err != nil {
		return nil, err
	}
	o := &clientOptions{}
	for _, opt := range opts {
		opt.apply(o)
	}
	return &Client{ec: ec, o: o}, nil
}

func (it *Client) Emit(subject string, e *pb.Event, opts ...EmitOption) error {
	return it.ec.Publish(subject, e)
}

func (it *Client) GetInterceptor() ClientEventInterceptor {
	if it.o.interceptor == nil {
		return EmptyClientInterceptor
	}
	return it.o.interceptor
}
