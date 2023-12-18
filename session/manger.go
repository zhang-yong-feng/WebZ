package session

import (
	"github.com/google/uuid"
	"github.com/zhang-yong-feng/webz"
)

// Manager 纯属用户体验好
type Manager struct {
	Propagator
	Store
}

// GetSession 将ctx里边拿取session
// 拿取后缓存到UserValues里边
func (m *Manager) GetSession(ctx *webz.Context) (Session, error) {
	sessId, err := m.Extract(ctx.Req)
	if err != nil {
		return nil, err
	}
	return m.Get(ctx.Req.Context(), sessId)
}

// InitSession 初始化session,并且注入http里response里边
func (m *Manager) InitSession(ctx *webz.Context) (Session, error) { //初始化
	id := uuid.New().String()
	sess, err := m.Generate(ctx.Req.Context(), id)
	if err != nil {
		return nil, err
	}
	err = m.Inject(id, ctx.Resp) //将响应注入到http中
	return sess, err
}

// RemoveSession 刷新session
func (m *Manager) RemoveSession(ctx *webz.Context) error {
	sess, err := m.GetSession(ctx)
	if err != nil {
		return err
	}
	err = m.Store.Remove(ctx.Req.Context(), sess.ID())
	if err != nil {
		return err
	}
	return m.Propagator.Remove(ctx.Resp)
}

// RefreshSession 删除session
func (m *Manager) RefreshSession(ctx *webz.Context) error {
	sess, err := m.GetSession(ctx)
	if err != nil {
		return err
	}
	return m.Refresh(ctx.Req.Context(), sess.ID())
}
