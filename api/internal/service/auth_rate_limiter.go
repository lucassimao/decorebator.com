package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type AuthLimitOperation string
type AuthLimitDimension string

const (
	AuthLimitSignup         AuthLimitOperation = "signup"
	AuthLimitLogin          AuthLimitOperation = "login"
	AuthLimitResetRequest   AuthLimitOperation = "reset_request"
	AuthLimitResetConsume   AuthLimitOperation = "reset_consume"
	AuthLimitPushUnregister AuthLimitOperation = "push_unregister"

	AuthLimitSource  AuthLimitDimension = "source"
	AuthLimitAccount AuthLimitDimension = "account"

	authLocalLimitKeys = 50000
)

type authLimitPolicy struct {
	limit  int64
	window time.Duration
}

type authLimitBucket struct {
	count     int64
	expiresAt time.Time
}

type AuthLimitDecision struct {
	Allowed    bool
	RetryAfter time.Duration
}

type AuthRateLimiter struct {
	redis        *redis.Client
	requireRedis bool
	namespace    string
	now          func() time.Time
	mu           sync.Mutex
	local        map[string]authLimitBucket
}

var authRedisLimitScript = redis.NewScript(`
local count = redis.call('INCR', KEYS[1])
if count == 1 then
  redis.call('PEXPIRE', KEYS[1], ARGV[1])
end
return {count, redis.call('PTTL', KEYS[1])}
`)

var authRedisReleaseScript = redis.NewScript(`
local count = tonumber(redis.call('GET', KEYS[1]) or '0')
if count <= 1 then
  redis.call('DEL', KEYS[1])
  return 0
end
return redis.call('DECR', KEYS[1])
`)

func NewAuthRateLimiter(
	redisClient *redis.Client,
	requireRedis bool,
	environment string,
	now func() time.Time,
) (*AuthRateLimiter, error) {
	if requireRedis && redisClient == nil {
		return nil, fmt.Errorf("configured Redis auth limiter is unavailable")
	}
	if now == nil {
		now = time.Now
	}
	namespace := "v1:" + environment
	if environment == "test" {
		namespace = uuid.NewString()
	}
	return &AuthRateLimiter{
		redis: redisClient, requireRedis: requireRedis, namespace: namespace,
		now: now, local: make(map[string]authLimitBucket),
	}, nil
}

func (l *AuthRateLimiter) Check(
	ctx context.Context,
	operation AuthLimitOperation,
	dimension AuthLimitDimension,
	value string,
) (AuthLimitDecision, error) {
	if l == nil || value == "" {
		return AuthLimitDecision{}, fmt.Errorf("auth limiter input is required")
	}
	policy, ok := authLimitPolicyFor(operation, dimension)
	if !ok {
		return AuthLimitDecision{}, fmt.Errorf("unknown auth limit policy")
	}
	key := l.key(operation, dimension, value)
	if l.redis != nil {
		result, err := authRedisLimitScript.Run(
			ctx, l.redis, []string{key}, policy.window.Milliseconds(),
		).Int64Slice()
		if err != nil || len(result) != 2 {
			return AuthLimitDecision{}, fmt.Errorf("auth limiter unavailable")
		}
		return authLimitDecision(result[0], time.Duration(result[1])*time.Millisecond, policy), nil
	}
	if l.requireRedis {
		return AuthLimitDecision{}, fmt.Errorf("auth limiter unavailable")
	}
	return l.checkLocal(key, policy), nil
}

// Release refunds one accepted reservation when authentication could not be
// evaluated because infrastructure failed.
func (l *AuthRateLimiter) Release(
	ctx context.Context,
	operation AuthLimitOperation,
	dimension AuthLimitDimension,
	value string,
) error {
	if l == nil || value == "" {
		return fmt.Errorf("auth limiter input is required")
	}
	if _, ok := authLimitPolicyFor(operation, dimension); !ok {
		return fmt.Errorf("unknown auth limit policy")
	}
	key := l.key(operation, dimension, value)
	if l.redis != nil {
		if _, err := authRedisReleaseScript.Run(ctx, l.redis, []string{key}).Int64(); err != nil {
			return fmt.Errorf("auth limiter unavailable")
		}
		return nil
	}
	if l.requireRedis {
		return fmt.Errorf("auth limiter unavailable")
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	bucket, exists := l.local[key]
	if !exists || bucket.count <= 1 {
		delete(l.local, key)
		return nil
	}
	bucket.count--
	l.local[key] = bucket
	return nil
}

// Clear removes the account bucket after successful authentication so normal
// sign-ins do not accumulate toward the failed-attempt ceiling.
func (l *AuthRateLimiter) Clear(
	ctx context.Context,
	operation AuthLimitOperation,
	dimension AuthLimitDimension,
	value string,
) error {
	if l == nil || value == "" {
		return fmt.Errorf("auth limiter input is required")
	}
	if _, ok := authLimitPolicyFor(operation, dimension); !ok {
		return fmt.Errorf("unknown auth limit policy")
	}
	key := l.key(operation, dimension, value)
	if l.redis != nil {
		if err := l.redis.Del(ctx, key).Err(); err != nil {
			return fmt.Errorf("auth limiter unavailable")
		}
		return nil
	}
	if l.requireRedis {
		return fmt.Errorf("auth limiter unavailable")
	}
	l.mu.Lock()
	delete(l.local, key)
	l.mu.Unlock()
	return nil
}

func (l *AuthRateLimiter) key(
	operation AuthLimitOperation,
	dimension AuthLimitDimension,
	value string,
) string {
	digest := sha256.Sum256([]byte(value))
	return "auth_limit:" + l.namespace + ":" + string(operation) + ":" +
		string(dimension) + ":" + hex.EncodeToString(digest[:])
}

func (l *AuthRateLimiter) checkLocal(key string, policy authLimitPolicy) AuthLimitDecision {
	now := l.now().UTC()
	l.mu.Lock()
	defer l.mu.Unlock()
	bucket := l.local[key]
	if bucket.expiresAt.IsZero() || !bucket.expiresAt.After(now) {
		bucket = authLimitBucket{expiresAt: now.Add(policy.window)}
	}
	if _, exists := l.local[key]; !exists && len(l.local) >= authLocalLimitKeys {
		for existingKey, existing := range l.local {
			if !existing.expiresAt.After(now) {
				delete(l.local, existingKey)
			}
		}
		if len(l.local) >= authLocalLimitKeys {
			return AuthLimitDecision{RetryAfter: policy.window}
		}
	}
	bucket.count++
	l.local[key] = bucket
	return authLimitDecision(bucket.count, bucket.expiresAt.Sub(now), policy)
}

func authLimitDecision(count int64, retryAfter time.Duration, policy authLimitPolicy) AuthLimitDecision {
	if count <= policy.limit {
		return AuthLimitDecision{Allowed: true}
	}
	if retryAfter <= 0 || retryAfter > policy.window {
		retryAfter = policy.window
	}
	return AuthLimitDecision{RetryAfter: retryAfter}
}

func authLimitPolicyFor(operation AuthLimitOperation, dimension AuthLimitDimension) (authLimitPolicy, bool) {
	policies := map[string]authLimitPolicy{
		string(AuthLimitSignup) + ":" + string(AuthLimitSource):         {limit: 100, window: time.Hour},
		string(AuthLimitSignup) + ":" + string(AuthLimitAccount):        {limit: 3, window: time.Hour},
		string(AuthLimitLogin) + ":" + string(AuthLimitSource):          {limit: 30, window: 10 * time.Minute},
		string(AuthLimitLogin) + ":" + string(AuthLimitAccount):         {limit: 5, window: 10 * time.Minute},
		string(AuthLimitResetRequest) + ":" + string(AuthLimitSource):   {limit: 10, window: time.Hour},
		string(AuthLimitResetRequest) + ":" + string(AuthLimitAccount):  {limit: 3, window: time.Hour},
		string(AuthLimitResetConsume) + ":" + string(AuthLimitSource):   {limit: 10, window: time.Hour},
		string(AuthLimitResetConsume) + ":" + string(AuthLimitAccount):  {limit: 5, window: time.Hour},
		string(AuthLimitPushUnregister) + ":" + string(AuthLimitSource): {limit: 30, window: 10 * time.Minute},
	}
	policy, ok := policies[string(operation)+":"+string(dimension)]
	return policy, ok
}

func RetryAfterSeconds(duration time.Duration) string {
	seconds := int64((duration + time.Second - 1) / time.Second)
	if seconds < 1 {
		seconds = 1
	}
	return strconv.FormatInt(seconds, 10)
}
