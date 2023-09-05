package my_redis_lock

import (
	"context"
	"my_redis_lock/repo"
	"sync"
	"testing"
)

const (
	addr   = "xx:xx"
	passwd = "xx"
)

func Test_NonBlockLock(t *testing.T) {
	client := repo.NewRedisClient("tcp", addr, passwd)
	lock1 := NewRedisLock(client, "test_key", WithExpireSeconds(5))
	lock2 := NewRedisLock(client, "test_key", WithExpireSeconds(8))

	ctx := context.Background()

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()

		if err := lock1.Lock(ctx); err != nil {
			t.Error(err)
			return
		}
	}()
	go func() {
		defer wg.Done()

		if err := lock2.Lock(ctx); err != nil {
			t.Error(err)
			return
		}
	}()

	wg.Wait()
}

func Test_BlockLock(t *testing.T) {
	client := repo.NewRedisClient("tcp", addr, passwd)
	lock1 := NewRedisLock(client, "test_key", WithBlock(), WithExpireSeconds(5))
	lock2 := NewRedisLock(client, "test_key", WithBlock(), WithExpireSeconds(8))

	ctx := context.Background()

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()

		if err := lock1.Lock(ctx); err != nil {
			t.Error(err)
			return
		}
	}()
	go func() {
		defer wg.Done()

		if err := lock2.Lock(ctx); err != nil {
			t.Error(err)
			return
		}
	}()

	wg.Wait()
}
