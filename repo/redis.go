package repo

import (
	"context"
	"errors"
	"github.com/gomodule/redigo/redis"
	"strings"
	"time"
)

// RedisClient redis客户端
type RedisClient struct {
	ClientOptions
	pool *redis.Pool
}

// NewRedisClient 新建redis client
func NewRedisClient(netWork, address, appKey string, opts ...ClientOption) *RedisClient {
	redisClient := &RedisClient{
		ClientOptions: ClientOptions{
			netWork: netWork,
			address: address,
			appKey:  appKey,
		},
	}

	for _, opt := range opts {
		opt(&redisClient.ClientOptions)
	}

	modifyClientOptions(&redisClient.ClientOptions)
	redisClient.pool = redisClient.initPool()

	return redisClient
}

// 初始化连接池
func (r *RedisClient) initPool() *redis.Pool {
	return &redis.Pool{
		// 池中最大活动（同时使用）的连接数
		MaxActive: r.maxActive,
		// 池中最大空闲连接数
		MaxIdle: r.maxIdle,
		// 当连接池耗尽时，是否等待可用连接
		Wait: r.wait,
		// 连接在池中空闲的最长时间
		IdleTimeout: time.Duration(r.idleTimeout) * time.Second,
		// 连接在池中的最长生命周期
		MaxConnLifetime: time.Duration(r.maxConnLifeTime) * time.Second,
		// 创建和返回一个新的 Redis 连接
		Dial: func() (redis.Conn, error) {
			return r.getRedisConn()
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

// 获取redis连接
func (r *RedisClient) getRedisConn() (redis.Conn, error) {
	if r.address == "" {
		panic("Cannot get redis address from config")
	}

	var dialOpts []redis.DialOption
	if len(r.appKey) > 0 {
		dialOpts = append(dialOpts, redis.DialPassword(r.appKey))
	}
	conn, err := redis.DialContext(context.Background(),
		r.netWork, r.address, dialOpts...)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (r *RedisClient) SETNEX(ctx context.Context, key, value string, expireSeconds int64) (int64, error) {
	if key == "" || value == "" {
		return -1, errors.New("SETNEX key or value is nil")
	}

	conn, err := r.pool.GetContext(ctx)
	if err != nil {
		return -1, err
	}
	defer conn.Close()

	reply, err := conn.Do("SET", key, value, "EX", expireSeconds, "NX")
	if err != nil {
		return -1, err
	}

	if replyStr, ok := reply.(string); ok && strings.ToLower(replyStr) == "ok" {
		return 1, nil
	}

	return redis.Int64(reply, err)
}

func (r *RedisClient) EVAL(ctx context.Context, src string, keyCount int,
	keyAndArgs []interface{}) (interface{}, error) {

	args := make([]interface{}, len(keyAndArgs)+2)
	args[0] = src
	args[1] = keyCount
	copy(args[2:], keyAndArgs)

	conn, err := r.pool.GetContext(ctx)
	if err != nil {
		return -1, err
	}
	defer conn.Close()

	return conn.Do("EVAL", args)
}
