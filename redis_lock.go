package my_redis_lock

import (
	"context"
	"errors"
	"fmt"
	"my_redis_lock/repo"
	"sync/atomic"
	"time"
)

var ErrNil = "redigo: nil returned"

var ErrLockAcquiredByOthers = errors.New("lock is acquired by others")

const (
	RedisLockPrefix = "REDIS_LOCK_PREFIX_"
)

// RedisLock redis lock
type RedisLock struct {
	// 分布式锁配置选项
	RedisLockOptions
	// redis客户端
	client *repo.RedisClient
	// 分布式锁key
	key string
	// token,保证每个协程的token唯一,用于校验锁的所属关系
	token string

	// 看门狗运行标记
	watchDogRunning int32
	// 停止看门狗
	stopWatchDog context.CancelFunc
}

// NewRedisLock 新建redis lock
func NewRedisLock(client *repo.RedisClient, key string, opts ...RedisLockOption) *RedisLock {
	redisLock := &RedisLock{
		client: client,
		key:    RedisLockPrefix + key,
		token:  GetProcessAndGoroutineIDStr(),
	}

	for _, opt := range opts {
		opt(&redisLock.RedisLockOptions)
	}

	modifyRedisLockOptions(&redisLock.RedisLockOptions)

	return redisLock
}

// Lock 锁定
func (r *RedisLock) Lock(ctx context.Context) (err error) {
	defer func() {
		if err != nil {
			return
		}

		// 启动看门狗
		r.watchDog(ctx)
	}()
	err = r.tryLock(ctx)
	if err == nil {
		return
	}

	// 非阻塞模式,直接返回错误
	if !r.isBlock {
		return
	}

	// 阻塞模式,根据错误类型判断是否可进行重试
	retry := r.canRetry(err)
	if !retry {
		// 错误不可重试
		return
	}
	err = r.blockingLock(ctx)
	return
}

// 启用看门狗
func (r *RedisLock) watchDog(ctx context.Context) {
	// 若未启用看门狗,直接返回
	if !r.isWatchDogMode {
		return
	}

	// 确保之前的看门狗已经被回收
	for atomic.CompareAndSwapInt32(&r.watchDogRunning, 0, 1) {
	}

	ctx, r.stopWatchDog = context.WithCancel(ctx)

	go func() {
		defer atomic.StoreInt32(&r.watchDogRunning, 0)

		r.runWatchDog(ctx)
	}()

}

//
func (r *RedisLock) runWatchDog(ctx context.Context) {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		select {
		case <-ctx.Done():
			return
		default:
		}

		_ = r.delayExpire(ctx, r.expireSeconds)
	}
}

// 延期过期时间
func (r *RedisLock) delayExpire(ctx context.Context, expireSecond int64) error {
	keyAndArgs := []interface{}{r.key, r.token, expireSecond + 5}
	reply, err := r.client.EVAL(ctx, LuaDelayDistributionLockExpire, 1, keyAndArgs)
	if err != nil {
		return err
	}

	rep, _ := reply.(int64)
	if rep != 1 {
		return errors.New("can not expire lock without ownership of lock")
	}
	return nil
}

// 尝试拿锁
func (r *RedisLock) tryLock(ctx context.Context) error {
	reply, err := r.client.SETNEX(ctx, r.key, r.token, r.expireSeconds)
	if err != nil {
		return err
	}

	if reply == 1 {
		return nil
	}

	// reply != 1,说明已经被其他协程拿到了锁
	return fmt.Errorf("reply :%d.err :%w", reply, ErrLockAcquiredByOthers)
}

// 根据错误类型校验是否可以重试
func (r *RedisLock) canRetry(err error) bool {
	return errors.Is(err, ErrLockAcquiredByOthers) || err.Error() == ErrNil
}

// 阻塞拿锁
func (r *RedisLock) blockingLock(ctx context.Context) error {
	deadlineCh := time.After(time.Duration(r.blockWaitingSeconds) * time.Second)
	// 50ms 轮询
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		select {
		case <-deadlineCh:
			// 轮询超时
			return fmt.Errorf("blockingLock timeout!err: %w", ErrLockAcquiredByOthers)
		case <-ctx.Done():
			// ctx 终止
			return fmt.Errorf("ctx end!err: %w", ctx.Err())
		default:
			// 继续执行
		}

		err := r.tryLock(ctx)
		if err == nil {
			// 加锁成功,直接返回
			return nil
		}
		if !r.canRetry(err) {
			// 发生了不可重试的错误,直接返回
			return err
		}
	}
	return nil
}

// UnLock 解锁
func (r *RedisLock) UnLock(ctx context.Context) error {
	defer func() {
		// 停止看门狗
		if r.stopWatchDog != nil {
			r.stopWatchDog()
		}
	}()

	keysAndArgs := []interface{}{r.key, r.token}
	reply, err := r.client.EVAL(ctx, LuaCheckAndDelDistributionLock, 1, keysAndArgs)
	if err != nil {
		return err
	}

	if ret, _ := reply.(int64); ret != 1 {
		return errors.New("can not unlock without ownership of lock")
	}

	return nil
}
