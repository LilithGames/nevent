package nevent

import (
	"context"

	pb "github.com/LilithGames/nevent/proto"
	"github.com/nats-io/nats.go"
)

type ServerEventHandler func(ctx context.Context, t pb.EventType, m *nats.Msg) (interface{}, error)
type ServerInterceptor func(next ServerEventHandler) ServerEventHandler

func IdentityServerInterceptor() ServerInterceptor {
	return func(next ServerEventHandler) ServerEventHandler {
		return func(ctx context.Context, t pb.EventType, m *nats.Msg) (interface{}, error) {
			return next(ctx, t, m)
		}
	}
}

func FuncServerInterceptor(i ServerInterceptor) ServerOption {
	return newFuncServerOption(func(o *serverOptions) {
		if o.interceptor != nil {
			panic("interceptor already exists")
		}
		o.interceptor = i
	})
}

func ChainServerInterceptor(items ...ServerInterceptor) ServerInterceptor {
	return func(next ServerEventHandler) ServerEventHandler {
		current := next
		for i := len(items) - 1; i >= 0; i-- {
			current = items[i](current)
		}
		return current
	}
}

type ClientEventInvoker func(ctx context.Context, t pb.EventType, m *nats.Msg) (interface{}, error)
type ClientInterceptor func(next ClientEventInvoker) ClientEventInvoker

func IdentityClientInterceptor() ClientInterceptor {
	return func(next ClientEventInvoker) ClientEventInvoker {
		return func(ctx context.Context, t pb.EventType, m *nats.Msg) (interface{}, error) {
			return next(ctx, t, m)
		}
	}
}

func FuncClientInterceptor(i ClientInterceptor) ClientOption {
	return newFuncClientOption(func(o *clientOptions) {
		o.interceptor = i
	})
}

func ChainClientInterceptor(items ...ClientInterceptor) ClientInterceptor {
	return func(next ClientEventInvoker) ClientEventInvoker {
		current := next
		for i := len(items) - 1; i >= 0; i-- {
			current = items[i](current)
		}
		return current
	}
}
