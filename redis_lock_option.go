package my_redis_lock

type RedisLockOptions struct {
	// 锁过期时间
	expireSeconds int64
	// 是否是阻塞模式
	isBlock bool
	// 阻塞模式下阻塞的最长时间
	blockWaitingSeconds int64
	// 是否启用看门狗
	isWatchDogMode bool
}

type RedisLockOption func(*RedisLockOptions)

func modifyRedisLockOptions(r *RedisLockOptions) {
	if r.isBlock && r.blockWaitingSeconds <= 0 {
		// 阻塞模式下，阻塞时间上限默认5s
		r.blockWaitingSeconds = 5
	}

	if r.expireSeconds > 0 {
		return
	}

	// 若未设置锁的过期时间，默认启用看门狗
	r.isWatchDogMode = true
	r.expireSeconds = 30
}

func WithExpireSeconds(expireSeconds int64) RedisLockOption {
	return func(r *RedisLockOptions) {
		r.expireSeconds = expireSeconds
	}
}

func WithBlock() RedisLockOption {
	return func(r *RedisLockOptions) {
		r.isBlock = true
	}
}

func WithBlockWaitingSeconds(blockWaitingSeconds int64) RedisLockOption {
	return func(r *RedisLockOptions) {
		r.blockWaitingSeconds = blockWaitingSeconds
	}
}

func WithWatchDog() RedisLockOption {
	return func(r *RedisLockOptions) {
		r.isWatchDogMode = true
	}
}
