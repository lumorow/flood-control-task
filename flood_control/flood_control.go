package flood_control

import (
	"context"
	"fmt"
	"github.com/go-redis/redis"
	"strconv"
	"time"
)

type FloodControl struct {
	redisClient *redis.Client
	timeSize    int
	maxRequests int
}

func NewFloodControl(redisClient *redis.Client, windowSize, maxRequests int) *FloodControl {
	return &FloodControl{redisClient, windowSize, maxRequests}
}

func (fc *FloodControl) Check(ctx context.Context, userID int64) (bool, error) {
	key := fmt.Sprintf("floodcontrol:%d", userID)

	now := time.Now().Unix()

	fc.redisClient.ZRemRangeByScore(key, "0", strconv.FormatInt(now-int64(fc.timeSize), 10))

	fc.redisClient.ZAdd(key, redis.Z{Score: float64(now), Member: now})

	count, err := fc.redisClient.ZCard(key).Result()
	if err != nil {
		return false, err
	}

	return count <= int64(fc.maxRequests), nil
}
