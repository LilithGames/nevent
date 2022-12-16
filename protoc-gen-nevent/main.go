package main

import (
	pgs "github.com/lyft/protoc-gen-star"
	pgsgo "github.com/lyft/protoc-gen-star/lang/go"
	"google.golang.org/protobuf/types/pluginpb"
)

func main() {
	pbfeatures := uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
	g := pgs.Init(pgs.DebugEnv("NEVENT_DEBUG"), pgs.SupportedFeatures(&pbfeatures))
	g.RegisterModule(NewEvent())
	g.RegisterPostProcessor(pgsgo.GoFmt())
	g.Render()
}
