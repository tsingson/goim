package dao

import (
	"context"
	"time"

	"github.com/gomodule/redigo/redis"

	"github.com/Terry-Mao/goim/internal/logic/conf"
)

// PushMsg  interface for kafka / nats
type PushMsg interface {
	PublishMessage(topic, ackInbox string, key string, msg []byte) error
	Close() error
}

// Dao dao.
type Dao struct {
	c           *conf.Config
	push        PushMsg
	redis       *redis.Pool
	redisExpire int32
}

// New new a dao and return.
func New(c *conf.Config) *Dao {

	d := &Dao{
		c:           c,
		redis:       newRedis(c.Redis),
		redisExpire: int32(time.Duration(c.Redis.Expire) / time.Second),
	}

	if c.UseNats {
		d.push = NewNats(c)
	} else {
		d.push = NewKafka(c)
	}
	return d
}

func newRedis(c *conf.Redis) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     c.Idle,
		MaxActive:   c.Active,
		IdleTimeout: time.Duration(c.IdleTimeout),
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial(c.Network, c.Addr,
				redis.DialConnectTimeout(time.Duration(c.DialTimeout)),
				redis.DialReadTimeout(time.Duration(c.ReadTimeout)),
				redis.DialWriteTimeout(time.Duration(c.WriteTimeout)),
				redis.DialPassword(c.Auth),
			)
			if err != nil {
				return nil, err
			}
			return conn, nil
		},
	}
}

// Close close the resource.
func (d *Dao) Close() error {
	return d.redis.Close()
}

// Ping dao ping.
func (d *Dao) Ping(c context.Context) error {
	return d.pingRedis(c)
}
