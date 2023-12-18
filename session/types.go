package session

import (
	"context"
	"net/http"
)

// 管理Session 支持增删改查刷新等功能
type Store interface {

	//Generate 生成session
	Generate(ctx context.Context, id string) (Session, error)

	//Refresh  创建或者刷新
	Refresh(ctx context.Context, id string) error

	//Remove 删除
	Remove(ctx context.Context, id string) error

	//Get 获取
	Get(ctx context.Context, id string) (Session, error)
}

// Session 寻找和存储用户设置的数据
type Session interface {

	//Get 根据key拿取数据
	Get(ctx context.Context, key string) (any, error)

	//Set 设置键值对
	Set(ctx context.Context, key string, val string) error

	ID() string
}

// Propagator 抽象层将session存储到不同地方
type Propagator interface {

	//Inject 将session id 注入里边 必须幂等
	Inject(id string, writer http.ResponseWriter) error

	//Extrac 将session id 从http.Request里边提取出来
	Extract(req *http.Request) (string, error)

	//Remove 将session id从http.ResponseWriter中删除
	Remove(writer http.ResponseWriter) error
}
