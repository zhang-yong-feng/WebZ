package webz

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"sync"
)

const BodyBytesKey = "_webz/bodybyteskey"

// Context 代表上下文的抽象
// 没有用自带的http包里边的Context主要是想实现更多东西
type Context struct {
	Req         *http.Request
	Resp        http.ResponseWriter
	index       int8
	handles     HandlesChain
	FullPath    string            //完整处理器路径
	mu          sync.RWMutex      //锁保护线程安全的
	Keys        map[string]any    //key/value结构存储一些数据
	queryValue  url.Values        //这个是QueryValue这个函数没有缓存的问题
	PathParams  map[string]string //这个是解决:id路径参数问题//获取是在查询路由时获取//现在我还没有获取
	MathedRoute string            //拿取命中路由
	//下边这俩是为了设置返回值的//为什么不用resp这个，主要是因为会绕开RespData和Resptatuscode这两个//所以下边这俩主要是middleware读写用的
	RespStatusCode int
	RespData       []byte
	tplEngine      TemplateEngine //这个是渲染界面的
}

// Next控制流程
func (c *Context) Next() {
	c.index++
	for c.index < int8(len(c.handles)) {
		c.handles[c.index](c)
		c.index++
	}
}

// JSON
// 处理输出 将输入的数据转换一下然后提交给http
func (c *Context) JSON(code int, val any) error {
	bs, err := json.Marshal(val)
	if err != nil {
		return err
	}
	c.RespData = bs
	c.RespStatusCode = code
	return err
}

func (c *Context) JSONOK(val any) error {
	return c.JSON(http.StatusOK, val)
}

// SetCookie 发送cookie
func (c *Context) SetCookie(ck *http.Cookie) {
	http.SetCookie(c.Resp, ck)
}

// BindJSON 读取数据正常读取Request Body 只能读取一次
// 读取解析body里边的值放入val中
func (c *Context) BindJSON(val interface{}) error {
	if val == nil { //不能为空类型
		return errors.New("输入不能为nil")
	}
	decoder := json.NewDecoder(c.Req.Body) //解码
	return decoder.Decode(val)             //将值放入val中
}

// BindJSONOpt 控制是否选择 UseNumber DisallowUnknown
// UseNumber 这个选项用于将所有数字类型转换为 Number 类型。
// disallowUnknown 这个选项用于禁止未知属性。
func (c *Context) BindJSONOpt(val interface{}, useNumber, disallowUnknown bool) error {
	if val == nil { //不能为空类型
		return errors.New("输入不能为nil")
	}
	decoder := json.NewDecoder(c.Req.Body) //解码
	if useNumber {
		decoder.UseNumber()
	}
	if disallowUnknown {
		decoder.DisallowUnknownFields()
	}
	return decoder.Decode(val) //将值放入val中
}

// MuchBindJSON  多次读取数据
func (c *Context) MuchBindJSON(val interface{}) (err error) {
	if val == nil { //不能为空类型
		return errors.New("输入不能为nil")
	}
	var body []byte
	if cb, ok := c.get(BodyBytesKey); ok {
		if cbb, ok := cb.([]byte); ok {
			body = cbb
		}
	}
	if body == nil {
		body, err = io.ReadAll(c.Req.Body)
		if err != nil {
			return err
		}
		c.set(BodyBytesKey, body)
	}
	return json.Unmarshal(body, val)
}

// get 线程安全拿取context里边keys的数据
func (c *Context) get(key string) (value any, exists bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, exists = c.Keys[key]
	return
}

// set 线程安全放入context里边keys的数据
func (c *Context) set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Keys == nil {
		c.Keys = make(map[string]any)
	}

	c.Keys[key] = value
}

// FormValue 处理表单数据Form
func (c *Context) FormValue(key string) (string, error) {
	// return c.Req.FormValue(key),nil可以直接这样写，但是这样写就没有error的问题了，它会直接帮助我们解决这个问题，我们并不希望它来帮助我们解决这个问题
	err := c.Req.ParseForm() //重复调用会不会产生问题，因为这个调用是幂等的
	if err != nil {
		return "", err
	}
	return c.Req.FormValue(key), nil
}

// QueryValue 查询参数带问号
func (c *Context) QueryValue(key string) (string, error) {
	// return c.Req.URL.Query().Get(key),nil//就是这个现在我可以这样写
	//但是这样写有个问题就是我的Query不像ParseForm一样有缓存//就是它每次都make,所以其实没有必要这样，我们可以给他加上缓存
	//如何加上缓存//首先我在context那边加上有个新的queryValue//然后加上判断
	if c.queryValue == nil {
		c.queryValue = c.Req.URL.Query()
	}
	vals, ok := c.queryValue[key]
	if !ok {
		return "", errors.New("key不存在")
	}
	return vals[0], nil
}

func (c *Context) PathValue(key string) (string, error) {
	val, ok := c.PathParams[key]
	if !ok {
		return "", errors.New("web: key不存在")
	}
	return val, nil
}

func (c *Context) Render(tplName string, data any) error {
	var err error
	c.RespData, err = c.tplEngine.Render(c.Req.Context(), tplName, data)
	if err != nil {
		c.RespStatusCode = http.StatusInternalServerError
		return err
	}
	c.RespStatusCode = http.StatusOK
	return nil
}
