package webz

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

//这里我会写一些简单或者复杂的路由进行测试查看树是否建对

func TestRouter1(t *testing.T) {
	//构造路由树
	testTrees := []struct {
		method string
		path   string
	}{
		{
			method: http.MethodGet,
			path:   "/a",
		},
		{
			method: http.MethodGet,
			path:   "/aa",
		},
	}
	var mockHandle HandlesChain
	mockHandle = append(mockHandle, func(ctx *Context) {})
	webz := NewHTTPServer()
	for _, route := range testTrees {
		webz.addRoute(route.method, route.path, mockHandle)
	}
	//验证路由器是否一样
	treeok := webz.trees[0].root
	wanttree := &node{
		path:         "/a",
		indices:      "a",
		fullPath:     "/a",
		handlesChain: mockHandle,
	}
	wanttree.children = append(wanttree.children, &node{
		path:         "a",
		fullPath:     "/aa",
		handlesChain: mockHandle,
	})
	//断言两个是否相等
	msg, ok := wanttree.equal(treeok)
	fmt.Println(msg, ok)
}

func (n *node) equal(y *node) (string, bool) {
	if n.path != y.path {
		return "path不相等", false
	}
	if len(n.children) != len(y.children) {
		return "children数量不相等", false
	}
	if n.fullPath != y.fullPath {
		return "fullpath不相等", false
	}
	if n.indices != y.indices {
		return "indices不相等", false
	}
	//比较handler//用反射比
	for _, iiii := range n.handlesChain {
		nhandler := reflect.ValueOf(iiii)
		yhandler := reflect.ValueOf(iiii)
		if nhandler != yhandler {
			return "handlechain不相等", false
		}
	}
	for ii, c := range n.children {
		dst := y.children[ii]
		msg, ok := c.equal(dst)
		if !ok {
			return msg, false
		}
	}
	return "相等", true
}
