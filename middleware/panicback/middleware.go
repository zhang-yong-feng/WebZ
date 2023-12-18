package panicback

import "github.com/zhang-yong-feng/webz"

type MiddlewareBuilder struct {
	StatusCode int
	Data       []byte
}

func (m MiddlewareBuilder) Build() webz.HandleFunc {
	return func(ctx *webz.Context) {
		defer func() {
			if err := recover(); err != nil {
				ctx.RespData = m.Data
				ctx.RespStatusCode = m.StatusCode
			}

		}()
		ctx.Next()
	}
}
