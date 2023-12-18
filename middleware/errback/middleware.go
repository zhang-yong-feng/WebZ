package errback

import "github.com/zhang-yong-feng/webz"

type MiddlewareBuilder struct {
	//只能返回固定的值
	//不能做到动态渲染
	resp map[int][]byte
}

// NewMiddlewareBuilder 初始化创建MiddlewareBuilder实例
func NewMiddlewareBuilder() *MiddlewareBuilder {
	return &MiddlewareBuilder{
		resp: make(map[int][]byte)}
}

// AddCode 添加数据
func (m *MiddlewareBuilder) AddCode(status int, data []byte) *MiddlewareBuilder {
	m.resp[status] = data
	return m
}

func (m MiddlewareBuilder) Build() webz.HandleFunc {
	return func(ctx *webz.Context) {
		ctx.Next()
		resp, ok := m.resp[ctx.RespStatusCode]
		if ok {
			//直接篡改
			ctx.RespData = resp
		}
	}
}
