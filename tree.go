package webz

import (
	"net/http"
)

type tree struct {

	//method 用来记录方法
	//现在我只支持get//post
	method string

	//root 头节点
	//初始化赋值并且给它初始化为"/""
	root *node
}

type node struct {

	//children 子节点
	children []*node

	////当前相对路径
	path string

	// indices 节点与子节点分开的第一个路径优化搜索
	indices string

	//fullPath 完整路径
	fullPath string

	//handlersChain 处理器函数结合
	handlesChain HandlesChain
}

// nodeBack 用来记录查找时的返回值
type nodeBack struct {

	//asteriskpath 记录*路径
	asteriskpath string
	//asteriskhandlesChain 记录*处理器
	asteriskhandlesChain HandlesChain

	//slashpath 记录/路径
	slashpath string
	//slashhandlesChain 记录/处理器
	slashhandlesChain HandlesChain

	//path 真正路径
	path string
	//handlesChain 真正处理器
	handlesChain HandlesChain
}

// newRouter 初始化路由树
func newRouter() []tree {
	var trees []tree
	trees = append(trees, tree{
		method: http.MethodGet,
		root:   &node{},
	})
	trees = append(trees, tree{
		method: http.MethodPost,
		root:   &node{},
	})
	return trees
}

func (n *node) addRoute(path string, handles []HandleFunc) {
	if path == "" {
		panic("path路径不能为空")
	}
	//假如为空
	if len(n.path) == 0 {
		//直接创建
		n.createNowNode(path, handles)
		return
	}
	n.addRouteSupplement(path, handles)

}

// addRouteSupplement 为addRoute提供循环函数所创
func (n *node) addRouteSupplement(path string, handles []HandleFunc) {
	//假如俩个路径相等
	if path == n.path {
		if len(n.handlesChain) != 0 {
			panic("每一个路径只能注册一次处理器")
		}
		n.handlesChain = handles
	}
	//先找n.path和path最长公共部分//不包含*和:
	publicpath := n.publicPath(path)
	if len(publicpath) < len(n.path) { //n.path要分裂
		//先把n.path分开 n.path原本的空间存储缩小的共同地址//新来的存储剩下的
		newNode := node{
			path:         n.path[len(publicpath):],
			fullPath:     n.fullPath,
			handlesChain: n.handlesChain,
			children:     n.children,
			indices:      n.indices,
		}
		n.indices = newNode.path[0:1]
		n.handlesChain = nil
		//n.children = append(n.children, &newNode)
		//上边写的不对//应该先赋空然后加上新的
		n.children = nil
		n.children = append(n.children, &newNode)
		//n.fullPath = n.fullPath[:len(publicpath)]
		//这个也不对 完整路径应该是减去后边的那一点
		n.fullPath = n.fullPath[:len(n.fullPath)-(len(n.path)-len(publicpath))]
		n.path = n.path[:len(publicpath)]
	}
	//先把剩下的path部分进行截取判断
	path = path[len(publicpath):]
	//判断接下来path对于indices是否还有重复
	var iii int
	for iii = 0; iii < len(n.indices); iii++ {
		if n.indices[iii] == path[0] {
			if n.indices[iii] == '*' || n.indices[iii] == ':' {
				continue
			}
			break
		}
	}
	if iii == len(n.indices) { //说明没有相同公共部分
		//查看是否有:或*和这个重复直接panic
		{
			if path[0] == '*' || path[0] == ':' {
				for iiiii := 0; iiiii < len(n.indices); iiiii++ {
					if n.indices[iiiii] == '*' || n.indices[iiiii] == ':' {
						panic("同一路径下不能同时有*和:")
					}
				}
			}
		}
		//如果添加的是/:d,/*则需要分开
		{
			if (path[0] == '/' && path[1] == ':') || (path[0] == '/' && path[1] == '*') {
				node111 := node{
					path:     "/",
					indices:  path[1:2],
					fullPath: n.fullPath + "/",
				}
				node222 := node{
					path:         path[1:],
					fullPath:     n.fullPath + path,
					handlesChain: handles,
				}
				node111.children = append(node111.children, &node222)
				n.indices = n.indices + "/"
				n.children = append(n.children, &node111)
				return
			}
		}
		//先将n.path搜索优化添加
		n.indices = n.indices + path[0:0+1]
		//现在path已经被截取好
		nodenew := node{
			path:         path,
			fullPath:     n.fullPath + path,
			handlesChain: handles,
		}
		//添加到n节点里边
		n.children = append(n.children, &nodenew)
	} else {
		//有相同公共部分，再次调用这个函数
		for _, child := range n.children {
			if child.path[0] == n.fullPath[iii] {
				child.addRoute(path, handles)
				break
			}
		}
	}
}

// createNowNode
// 用于addRoute直接创建//没有一个数据
func (n *node) createNowNode(path string, handles []HandleFunc) {
	//查找是否有:
	//这里我认为是否有*没有关系
	for a, c := range path {
		if c == ':' { //里边有: 在第a个
			n.path = path[:a]
			n.fullPath = n.path
			n.indices = path[a : a+1]
			child := node{
				path:         path[a:],
				fullPath:     path,
				handlesChain: handles,
			}
			n.children = []*node{&child}
			return
		}
	}
	n.path = path
	n.fullPath = path
	n.handlesChain = handles
}

// publicPath 寻找最长相同路径路径
func (n *node) publicPath(path string) string {
	//找到最小len
	min := len(path)
	if min > len(n.path) {
		min = len(n.path)
	}
	i := 0
	for ; i < min; i++ {
		if path[i] == ':' || path[i] == '*' {
			break
		}
		if path[i] != n.path[i] {
			break
		}
	}
	if i == 0 { //没有相同的
		return ""
	}
	return path[:i]
}

// findTree递归查找节点
func (n *node) findTree(path string, nodee *nodeBack) {
	//先两个是否相等
	if n.path == path {
		nodee.path = n.fullPath
		nodee.handlesChain = n.handlesChain
		return
	}

	if len(path) == 0 {
		return
	}

	//先判断前边n.path是否相等//不相等返回//先求出两个谁短
	min := len(n.path)
	if len(path) < min { //最大是n.path说明两个一定不相等
		return
	}
	if n.path != path[:min] { //前边相等的地方不相等
		return
	}

	//查找当前节点是否有:或者*
	//将这俩的node进行更新
	for _, b := range n.indices {
		if b == '*' {
			back, b := n.childrenPrice('*')
			if b {
				nodee.asteriskpath = back.fullPath
				nodee.asteriskhandlesChain = back.handlesChain
				break
			}
		}
		if b == ':' {
			back, b := n.childrenPrice(':')
			if b {
				nodee.slashpath = back.fullPath
				nodee.slashhandlesChain = back.handlesChain
				break
			}
		}
	}
	//查看是否能进入下一个阶段
	for a := range n.indices {
		//判断是否能进入下一阶段
		if n.indices[a] == path[min] {
			back, bb := n.childrenPrice(n.indices[a])
			if bb { //找到
				back.findTree(path[min:], nodee)
			}
		}
	}
}

// 已经有一个单个字符，返回n.children的子节点
func (n *node) childrenPrice(a byte) (*node, bool) {
	for i := 0; i < len(n.children); i++ {
		if n.children[i].path[0] == a {
			return n.children[i], true
		}
	}
	return nil, false
}
