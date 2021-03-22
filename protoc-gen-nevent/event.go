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
		"subject": func(m pgs.Message) string {
			options := m.Descriptor().GetOptions()
			opt, err := proto.GetExtension(options, pb.E_Subject)
			if err != nil {
				return ""
			}
			subject := opt.(*string)
			if subject == nil {
				return ""
			}
			return *subject
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
	if len(f.Services()) != 0 {
		it.Failf("syntax error: services not supported in event protobuf")
	}
	name := it.ctx.OutputPath(f).SetExt(".nevent.go").String()
	it.AddGeneratorTemplateFile(name, it.tpl, f)
}
