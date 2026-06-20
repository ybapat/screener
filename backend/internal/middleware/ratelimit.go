package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/ybapat/screener/backend/pkg/response"
)

// RateLimit limits authenticated users by user ID and path.
func RateLimit(rdb *redis.Client, limit int, window time.Duration) func(http.Handler) http.Handler {
	return rateLimitMiddleware(rdb, limit, window, func(r *http.Request) string {
		userID := GetUserID(r.Context())
		return fmt.Sprintf("ratelimit:%s:%s", r.URL.Path, userID.String())
	})
}

// RateLimitByIP limits requests by remote IP and path, for use on unauthenticated routes.
func RateLimitByIP(rdb *redis.Client, limit int, window time.Duration) func(http.Handler) http.Handler {
	return rateLimitMiddleware(rdb, limit, window, func(r *http.Request) string {
		return fmt.Sprintf("ratelimit:ip:%s:%s", r.URL.Path, r.RemoteAddr)
	})
}

func rateLimitMiddleware(rdb *redis.Client, limit int, window time.Duration, keyFn func(*http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := keyFn(r)

			ctx := context.Background()
			count, err := rdb.Incr(ctx, key).Result()
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			if count == 1 {
				rdb.Expire(ctx, key, window)
			}

			remaining := int64(limit) - count
			if remaining < 0 {
				remaining = 0
			}

			ttl, _ := rdb.TTL(ctx, key).Result()
			resetAt := time.Now().Add(ttl).Unix()

			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
			w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetAt, 10))

			if count > int64(limit) {
				response.ErrorMsg(w, http.StatusTooManyRequests, "rate limit exceeded")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
