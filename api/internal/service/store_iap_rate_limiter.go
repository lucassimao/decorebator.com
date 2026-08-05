package service

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	storeIAPLimitWindow      = time.Minute
	storeIAPLocalUsers       = 10000
	storeIAPVerifyOperation  = "verify"
	storeIAPRestoreOperation = "restore"
)

type storeIAPBucket struct {
	windowStart time.Time
	verify      int
	restore     int
}

type StoreIAPRequestLimiter struct {
	redis        *redis.Client
	requireRedis bool
	now          func() time.Time
	mu           sync.Mutex
	inFlight     map[int64]struct{}
	local        map[int64]storeIAPBucket
}

var storeIAPRedisLimitScript = redis.NewScript(`
local count = redis.call('INCR', KEYS[1])
if count == 1 then
  redis.call('PEXPIRE', KEYS[1], ARGV[1])
end
return count
`)

func NewStoreIAPRequestLimiter(
	redisClient *redis.Client,
	requireRedis bool,
	now func() time.Time,
) (*StoreIAPRequestLimiter, error) {
	if requireRedis && redisClient == nil {
		return nil, fmt.Errorf("configured Redis limiter is unavailable")
	}
	if now == nil {
		now = time.Now
	}
	return &StoreIAPRequestLimiter{
		redis: redisClient, requireRedis: requireRedis, now: now,
		inFlight: make(map[int64]struct{}), local: make(map[int64]storeIAPBucket),
	}, nil
}

// Acquire allows one provider-calling request per user in flight. The caller
// must invoke the returned release function exactly once when allowed is true.
func (l *StoreIAPRequestLimiter) Acquire(
	ctx context.Context,
	userID int64,
	operation string,
) (release func(), allowed bool, err error) {
	if userID <= 0 || (operation != storeIAPVerifyOperation && operation != storeIAPRestoreOperation) {
		return nil, false, fmt.Errorf("valid IAP limit input is required")
	}
	l.mu.Lock()
	if _, busy := l.inFlight[userID]; busy {
		l.mu.Unlock()
		return nil, false, nil
	}
	l.inFlight[userID] = struct{}{}
	l.mu.Unlock()
	release = func() {
		l.mu.Lock()
		delete(l.inFlight, userID)
		l.mu.Unlock()
	}

	limit := 10
	if operation == storeIAPRestoreOperation {
		limit = 5
	}
	if l.redis != nil {
		key := "store_iap:" + operation + ":" + strconv.FormatInt(userID, 10)
		count, redisErr := storeIAPRedisLimitScript.Run(
			ctx, l.redis, []string{key}, storeIAPLimitWindow.Milliseconds(),
		).Int64()
		if redisErr != nil {
			release()
			return nil, false, fmt.Errorf("store IAP limiter unavailable")
		}
		if count > int64(limit) {
			release()
			return nil, false, nil
		}
		return release, true, nil
	}
	if l.requireRedis {
		release()
		return nil, false, fmt.Errorf("store IAP limiter unavailable")
	}

	now := l.now().UTC()
	l.mu.Lock()
	bucket := l.local[userID]
	if bucket.windowStart.IsZero() || now.Sub(bucket.windowStart) >= storeIAPLimitWindow {
		bucket = storeIAPBucket{windowStart: now}
	}
	if len(l.local) >= storeIAPLocalUsers {
		if _, exists := l.local[userID]; !exists {
			for existingUser, existing := range l.local {
				if now.Sub(existing.windowStart) >= storeIAPLimitWindow {
					delete(l.local, existingUser)
				}
			}
		}
		if _, exists := l.local[userID]; !exists && len(l.local) >= storeIAPLocalUsers {
			l.mu.Unlock()
			release()
			return nil, false, nil
		}
	}
	count := &bucket.verify
	if operation == storeIAPRestoreOperation {
		count = &bucket.restore
	}
	*count++
	l.local[userID] = bucket
	if *count > limit {
		l.mu.Unlock()
		release()
		return nil, false, nil
	}
	l.mu.Unlock()
	return release, true, nil
}
