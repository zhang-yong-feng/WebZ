// server是一个代表服务器的抽象，主要提供注册路由的接口，还有http包到web框架的桥梁
package webz

import (
	"net"
	"net/http"
)

// HandleFunc 处理器函数
type HandleFunc func(ctx *Context)

// HandlesChain 处理器函数结合
type HandlesChain []HandleFunc

type HTTPServerOption func(server *HTTPServer)

// HTTPServer对服务器进行抽象 实现Server里边的方法
// 这个属性小写是因为我只想通过NewHttpServer进行创建路由，不打算把这个暴露在外边
type HTTPServer struct {
	//RouterGroup 用来实现路由的分组
	//刚开始是初始化记录的是"/"
	RouterGroup

	//router 用来存储路由
	trees []tree

	tplEngine TemplateEngine
}

// NewHttpServer 进行初始化创建路由
func NewHTTPServer() *HTTPServer {
	httpServer := &HTTPServer{
		RouterGroup: RouterGroup{
			basePath: "/",
			root:     true,
		},
		trees: newRouter(),
	}
	httpServer.RouterGroup.httpServer = httpServer
	return httpServer
}

// 用来实现context构建，路由匹配，业务逻辑
// 不是路由构建
func (h *HTTPServer) ServeHTTP(writer http.ResponseWriter, Request *http.Request) {
	//拿到一个context对象
	ctx := &Context{
		Req:   Request,
		Resp:  writer,
		index: -1,
	}
	h.serve(ctx)

	//现在我在这返回响应，不在JSON那边返回响应
	if ctx.RespStatusCode != 0 {
		ctx.Resp.WriteHeader(ctx.RespStatusCode)
	}
	ctx.Resp.Write(ctx.RespData)
}

// server 接下来查找路由
func (h *HTTPServer) serve(ctx *Context) {
	h.findTree(ctx)
}

// findTree
// serve 调用的函数用来实现查找路由
func (h *HTTPServer) findTree(ctx *Context) {
	path := ctx.Req.URL.Path
	method := ctx.Req.Method
	var c bool
	var tree1 *node
	for _, b := range h.trees {
		if b.method == method {
			tree1 = b.root
			c = true
			break
		}
	}
	if !c { //没有找到没有找到该方法
		//应该返回404
		return
	}
	back := nodeBack{}
	tree1.findTree(path, &back)
	//现在查找到底匹配到哪个路由

	if len(back.handlesChain) != 0 { //正确匹配
		ctx.handles = back.handlesChain
		ctx.FullPath = back.path
		ctx.Next()
		return
	}
	//统计是否匹配://这里只要比对他们/数量就行
	if len(back.slashhandlesChain) != 0 {
		a := 0
		for b, c := range back.slashpath {
			if c == ':' {
				a = b
				break
			}
		}
		d := true
		for _, c := range path[a+1:] {
			if c == '/' {
				d = false
				break
			}
		}
		if d {
			ctx.handles = back.slashhandlesChain
			ctx.FullPath = back.slashpath
			ctx.Next()
			return
		}
	}
	//最后看是否匹配*//否则没有匹配到数据
	if len(back.asteriskhandlesChain) != 0 {
		ctx.handles = back.asteriskhandlesChain
		ctx.FullPath = back.asteriskpath
		ctx.Next()
		return
	}
}

// Start是用来监听端口的
// 主要是模仿http.ListenAndServe这个监听端口来封装从而实现更多东西
func (h *HTTPServer) Start(add string) error {
	l, err := net.Listen("tcp", add) //监听端口
	if err != nil {
		return err
	}
	//Serve用来创建服务的连接
	//handler里边有一个serverHTTP类型的接口负责将服务树构建并且实现路由匹配执行业务逻辑
	return http.Serve(l, h) //创建服务连接
}

func (h *HTTPServer) addRoute(method string, path string, handles HandlesChain) {
	if path == "" {
		panic("path不能为空")
	}
	//先断言简单判断
	assert1(path[0] == '/', "path must begin with '/'")
	assert1(method != "", "HTTP method can not be empty")
	assert1(len(handles) > 0, "there must be at least one handler")
	root := h.get(method) //拿到树的头节点
	if root == nil {      //现在我只让使用get和post两个方法
		panic("不能注册没有的method")
	}
	//开始创建树
	root.addRoute(path, handles)

}

// get trees拿取node
func (h *HTTPServer) get(method string) *node {
	for _, tree := range h.trees {
		if tree.method == method {
			return tree.root
		}
	}
	return nil
}

// assert1 断言所用
func assert1(guard bool, text string) {
	if !guard {
		panic(text)
	}
}
