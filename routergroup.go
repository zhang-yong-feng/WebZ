package webz

import (
	"net/http"
)

type RouterGroup struct { //路由分组

	//basePath记录上边路径
	//刚开始初始化为/
	basePath string

	//Handles 当前组处理器
	Handles HandlesChain

	httpServer *HTTPServer

	root bool
}

// GET 允许传入多个处理器，具体实现稍后再说
func (group *RouterGroup) GET(path string, handles ...HandleFunc) {
	// 思路
	//1，将path路径补全//因为path是相对路径
	//2，将分组前添加的处理器全部加上
	//3，将补全后的path注册到路由树上
	newpath := group.completePath(path)
	newhandles := group.addHandle(handles)
	group.httpServer.addRoute(http.MethodGet, newpath, newhandles)
}

// 添加分组前的处理器
func (group *RouterGroup) addHandle(handles []HandleFunc) []HandleFunc {
	nowhandles := len(group.Handles) + len(handles)
	newhandles := make([]HandleFunc, nowhandles)
	copy(newhandles, group.Handles)
	copy(newhandles[len(group.Handles):], handles)
	return newhandles
}

// completePath 用于将路径补全
func (group *RouterGroup) completePath(path string) string {
	newpath := group.basePath
	if newpath == "/" { //如果是则去掉最后一个元素
		newpath = ""
	}
	newpath = newpath + path //这是得到的新路径
	return newpath

}

func (group *RouterGroup) Use(middleware ...HandleFunc) *HTTPServer {
	group.Handles = append(group.Handles, middleware...)
	return group.httpServer
}

func (group *RouterGroup) Group(relativePath string, handles ...HandleFunc) *RouterGroup {
	//这里要有判断 就是路径不能为什么什么的
	//复制Handlefunc
	finalSize := len(group.Handles) + len(handles)
	mergedHandlers := make(HandlesChain, finalSize)
	copy(mergedHandlers, group.Handles)
	copy(mergedHandlers[len(group.Handles):], handles)
	//复制路径
	pathhhh := group.basePath + relativePath
	return &RouterGroup{
		Handles:    mergedHandlers,
		basePath:   pathhhh,
		httpServer: group.httpServer,
	}
}
