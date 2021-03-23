package nevent

import (
	"strings"

	pb "github.com/LilithGames/nevent/proto"
	anypb "github.com/golang/protobuf/ptypes/any"
	"google.golang.org/grpc/metadata"
)

func MDFromEvent(e *pb.Event) metadata.MD {
	md := metadata.MD{}
	for k, val := range e.Metadata {
		key := strings.ToLower(k)
		md[key] = append(md[key], val.Items...)
	}
	return md
}

func NewEvent(md metadata.MD, data *anypb.Any) *pb.Event {
	meta := make(map[string]*pb.MDValue)
	for k, val := range md {
		key := strings.ToLower(k)
		meta[key] = &pb.MDValue{Items: val}
	}
	return &pb.Event{Metadata: meta, Data: data}
}
