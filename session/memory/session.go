package memory

import (
	"context"
	"errors"
	"sync"
	"time"

	cache "github.com/patrickmn/go-cache"
	"github.com/zhang-yong-feng/webz/session"
)

type Store struct {
	mutex      sync.RWMutex //RWMutex 提供了一种在并发编程中常用的读写锁机制，它允许多个 goroutine 同时读取数据，但在写入数据时只允许一个 goroutine 进行操作，并且这个操作是排他的。
	sessions   cache.Cache
	expiration time.Duration
}

type Session struct {
	values sync.Map
	id     string
}

// NewStore 设置过期时间
func NewStore(expiration time.Duration) *Store {
	return &Store{
		sessions:   *cache.New(expiration, time.Second),
		expiration: expiration,
	}
}

func (s *Store) Generate(ctx context.Context, id string) (session.Session, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	sess := &Session{
		id:     id,
		values: sync.Map{},
	}
	s.sessions.Set(id, sess, s.expiration)
	return sess, nil
}

func (s *Store) Refresh(ctx context.Context, id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	val, ok := s.sessions.Get(id)
	if !ok {
		return errors.New("该id对应的id不存在")
	}
	s.sessions.Set(id, val, s.expiration)
	return nil
}

func (s *Store) Remove(ctx context.Context, id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.sessions.Delete(id)
	return nil

}

func (s *Store) Get(ctx context.Context, id string) (session.Session, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	sess, ok := s.sessions.Get(id)
	if !ok {
		return nil, errors.New("找不到session")
	}
	return sess.(*Session), nil
}

func (s *Session) Get(ctx context.Context, key string) (any, error) {
	val, ok := s.values.Load(key)
	if ok {
		return nil, errors.New("session找不到key")
	}
	return val, nil
}
func (s *Session) Set(ctx context.Context, key string, val string) error {
	s.values.Store(key, val)
	return nil
}
func (s *Session) ID() string {
	return s.id //只读取 //所以没有必要加锁
}
