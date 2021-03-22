package nevent

import (
	"github.com/nats-io/nats.go"
	natspb "github.com/nats-io/nats.go/encoders/protobuf"
)

type serverOptions struct {
	group string
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

func ServerGroup(group string) ServerOption {
	return newFuncServerOption(func(o *serverOptions) {
		o.group = group
	})
}

type Server struct {
	ec *nats.EncodedConn
	o *serverOptions
}

// NewServer server with the same group will listen event round-robin
func NewServer(nc *nats.Conn, opts ...ServerOption) (*Server, error) {
	ec, err := nats.NewEncodedConn(nc, natspb.PROTOBUF_ENCODER)
	if err != nil {
		return nil, err
	}
	o := &serverOptions{}
	for _, opt := range opts {
		opt.apply(o)
	}
	return &Server{ec: ec, o: o}, nil
}

type listenOptions struct{
	group string
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

func ListenGroup(group string) ListenOption {
	return newFuncListenOption(func(o *listenOptions){
		o.group = group
	})
}

func (it *Server) ListenEvent(subject string, cb nats.Handler, opts ...ListenOption) (*nats.Subscription, error) {
	lo := &listenOptions{}
	for _, opt := range opts {
		opt.apply(lo)
	}
	var group string
	if lo.group != "" {
		group = lo.group
	} else {
		group = it.o.group
	}
	if group != "" {
		return it.ec.QueueSubscribe(subject, it.o.group, cb)
	} else {
		return it.ec.Subscribe(subject, cb)
	}
}
