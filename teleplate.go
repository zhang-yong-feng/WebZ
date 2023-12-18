package webz

import (
	"bytes"
	"context"
	"html/template"
)

type TemplateEngine interface {
	Render(ctx context.Context, tplName string, data any) ([]byte, error) //tplName 渲染引擎 data 渲染页面用的数据
}

type GoTemplateEngine struct {
	T *template.Template
}

func (g *GoTemplateEngine) Render(ctx context.Context, tplName string, data any) ([]byte, error) {
	bs := &bytes.Buffer{}
	err := g.T.ExecuteTemplate(bs, tplName, data)
	return bs.Bytes(), err
}

func (g *GoTemplateEngine) ParseGlob(pattern string) error { //这个就是我不用现成的这个，我自己封装
	var err error
	g.T, err = template.ParseGlob(pattern)
	return err
}

func ServerWithTemplateEngine(tplEngine TemplateEngine) HTTPServerOption { //用户可以选择初始化模板
	return func(server *HTTPServer) {
		server.tplEngine = tplEngine
	}
}
