package nevent

import (
	"context"
	// pb "github.com/LilithGames/nevent/proto"
	pbi "google.golang.org/protobuf/runtime/protoiface"
)

type ServerEventInterceptor func(ctx context.Context, subject string, reply interface{}, e pbi.MessageV1)

func EmptyServerInterceptor(ctx context.Context, subject string, reply interface{}, e pbi.MessageV1) {}

func ServerInterceptor(i ServerEventInterceptor) ServerOption {
	return newFuncServerOption(func(o *serverOptions) {
		if o.interceptor != nil {
			panic("interceptor already exists")
		}
		o.interceptor = i
	})
}

func ChainServerInterceptor(items ...ServerEventInterceptor) ServerEventInterceptor {
	return func(ctx context.Context, subject string, reply interface{}, e pbi.MessageV1) {
		for _, item := range items {
			item(ctx, subject, reply, e)
		}
	}
}


type ClientEventInterceptor func(ctx context.Context, subject string, method interface{}, e pbi.MessageV1)

func EmptyClientInterceptor(ctx context.Context, subject string, method interface{}, e pbi.MessageV1) {}

func ClientInterceptor(i ClientEventInterceptor) ClientOption {
	return newFuncClientOption(func (o *clientOptions) {
		o.interceptor = i
	})
}

func ChainClientInterceptor(items ...ClientEventInterceptor) ClientEventInterceptor {
	return func(ctx context.Context, subject string, method interface{}, e pbi.MessageV1) {
		for _, item := range items {
			item(ctx, subject, method, e)
		}
	}
}
