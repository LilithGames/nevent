package main

import (
	"text/template"

	pb "github.com/LilithGames/nevent/proto"
	"github.com/golang/protobuf/proto"
	pgs "github.com/lyft/protoc-gen-star"
	pgsgo "github.com/lyft/protoc-gen-star/lang/go"
)

type event struct {
	*pgs.ModuleBase
	ctx pgsgo.Context
	tpl *template.Template
}

func NewEvent() pgs.Module {
	return &event{ModuleBase: &pgs.ModuleBase{}}
}

func (it *event) InitContext(c pgs.BuildContext) {
	it.ModuleBase.InitContext(c)
	it.ctx = pgsgo.InitContext(c.Parameters())
	tpl := template.New("event").Funcs(map[string]interface{}{
		"package": it.ctx.PackageName,
		"name": it.ctx.Name,
		"default": func(value interface{}, defaultValue interface{}) interface{} {
			switch v := value.(type) {
			case string:
				if v == "" {
					return defaultValue
				}
			default:
				panic("default: unknown type")
			}
			return value
		},
		"options": func(node pgs.Node) *pb.EventOption {
			var opt interface{}
			var err error
			switch n := node.(type) {
			case pgs.File:
				opt, err = proto.GetExtension(n.Descriptor().GetOptions(), pb.E_Foptions)
			case pgs.Service:
				opt, err = proto.GetExtension(n.Descriptor().GetOptions(), pb.E_Soptions)
			case pgs.Method:
				opt, err = proto.GetExtension(n.Descriptor().GetOptions(), pb.E_Moptions)
			default:
				panic("node options not supported")
			}
			if err != nil {
				return new(pb.EventOption)
			}
			return opt.(*pb.EventOption)
		},
	})
	it.tpl = template.Must(tpl.Parse(tmpl))
}

func (it *event) Name() string {
	return "event"
}

func (it *event) Execute(targets map[string]pgs.File, pkgs map[string]pgs.Package) []pgs.Artifact {
	for _, f := range targets {
		it.generate(f)
	}
	return it.Artifacts()
}

func (it *event) generate(f pgs.File) {
	name := it.ctx.OutputPath(f).SetExt(".nevent.go").String()
	f.Services()
	it.AddGeneratorTemplateFile(name, it.tpl, f)
}
