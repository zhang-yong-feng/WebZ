package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zhang-yong-feng/webz/session"
)

// 需要存储的数据类似
//	sess id    key     value
// map[string]map[string]string

type Store struct {
	prefix     string //前缀
	client     redis.Cmdable
	expiration time.Duration
}

func NewStore(client redis.Cmdable) *Store {
	return &Store{
		expiration: time.Minute * 15,
		client:     client,
		prefix:     "sessid",
	}
}

func (s *Store) Generate(ctx context.Context, id string) (session.Session, error) {
	key := rediskey(s.prefix, id)
	_, err := s.client.HSet(ctx, key, id, id).Result()
	if err != nil {
		return nil, err
	}
	s.client.Expire(ctx, key, s.expiration)
	return &Session{id: id, client: s.client, key: rediskey(s.prefix, id)}, nil
}

func (s *Store) Refresh(ctx context.Context, id string) error {
	key := rediskey(s.prefix, id)
	ok, err := s.client.Expire(ctx, key, s.expiration).Result()

	if err != nil {
		return err
	}
	if !ok {
		return errors.New("session对应的id不存在")
	}
	return nil
}

func (s *Store) Remove(ctx context.Context, id string) error {
	key := rediskey(s.prefix, id)
	_, err := s.client.Del(ctx, key).Result()
	return err
}

func (s *Store) Get(ctx context.Context, id string) (session.Session, error) {
	key := rediskey(s.prefix, id)
	cnt, err := s.client.Exists(ctx, key).Result() //判断在不在
	if err != nil {
		return nil, err
	}
	if cnt != 1 {
		return nil, errors.New("session对应的id不存在")
	}
	return &Session{
		key:    key,
		id:     id,
		client: s.client,
	}, nil
}

type Session struct {
	prefix string
	client redis.Cmdable
	id     string
	key    string
}

func (s *Session) Get(ctx context.Context, key string) (any, error) {
	//刚开始要判断在或者不在  这个数据
	// k := rediskey(s.prefix, s.id)
	val, err := s.client.HGet(ctx, s.key, key).Result()
	return val, err
}
func (s *Session) Set(ctx context.Context, key string, val string) error {
	const lua = `
	if redis.call("exists",KEYS[1])
	then
	return redis.call("hset",KEY[1],ARGV[1],ARGV[2])
	else
	return -1
	end
	`
	k := rediskey(s.prefix, s.id)
	res, err := s.client.Eval(ctx, lua, []string{k}, key, val).Int()
	if err != nil {
		return err
	}
	if res < 0 {
		return errors.New("session对应的id不存在")
	}
	return nil
}
func (s *Session) ID() string {
	return s.id
}

func rediskey(prefix, id string) string {
	return fmt.Sprintf("%s-%s", prefix, id)
}
