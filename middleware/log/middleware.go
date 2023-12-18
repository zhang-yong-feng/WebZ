package middleware

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/zhang-yong-feng/webz"
)

// LoggerInformation 主要是为了记录每一次来的请求的结构体
type LoggerInformation struct {
	//Notprints哪些路径不要打印
	Notprints []string
	notprints map[string]struct{}
	//Output输出地点
	Output io.Writer
	//LogContent输出内容
	LogContent LogContentfunc
}

type LOGContext struct {
	Req *http.Request

	//现在时间
	TimeNow time.Time

	//执行时间
	TimeStamp time.Duration

	Path string

	//请求方法
	Method string

	//命中路由
	FullPath string

	//响应码
	StatusCode int
}

// LogContentfunc 接口类型
type LogContentfunc func(l LOGContext) string

// Logger 给用户写好的一个基础功能记录信息
func Logger() webz.HandleFunc {
	return loggerUniversal(LoggerInformation{})
}

func loggerUniversal(conf LoggerInformation) webz.HandleFunc {
	//先初始化
	l := LoggerInformation{}
	l.LogContent = conf.LogContent
	if l.LogContent == nil {
		l.LogContent = logSynthesis
		//初始化LogContent
		//调用写好的初始化函数

	}
	l.Output = conf.Output
	if l.Output == nil {
		//输出到控制台
		l.Output = os.Stdout
	}
	l.Notprints = conf.Notprints
	//现在出现个问题  就是Notprints 这个是[]string类型 后边查找不好找，所以我打算变成map[string]struct{这种类型
	if length := len(l.Notprints); length > 0 {
		l.notprints = make(map[string]struct{}, length)
		for _, path := range l.Notprints {
			l.notprints[path] = struct{}{}
		}
	}
	return func(ctx *webz.Context) {
		start := time.Now()      //开始时间
		path := ctx.Req.URL.Path //路径

		ctx.Next() //先执行里边的逻辑

		//查看是否为不记录的路径
		if _, ok := l.notprints[path]; !ok {
			information := LOGContext{
				Req:        ctx.Req,
				Path:       path,
				FullPath:   ctx.FullPath,
				StatusCode: ctx.RespStatusCode,
			}
			information.TimeNow = time.Now()
			information.TimeStamp = information.TimeNow.Sub(start)
			information.Method = ctx.Req.Method
			fmt.Fprint(l.Output, l.LogContent(information))
		}
	}
}

// defaultLogFormatter is the default log format function Logger middleware uses.
var logSynthesis = func(information LOGContext) string {

	if information.TimeStamp > time.Minute {
		information.TimeStamp = information.TimeStamp.Truncate(time.Second)
	}
	return fmt.Sprintf("[webz] %v| %3d |%v | %13v |%-7s %#v\n",
		information.TimeNow.Format("2006/01/02 - 15:04:05"),
		information.StatusCode,
		information.FullPath,
		information.TimeStamp,
		information.Method,
		information.Path,
	)
}
