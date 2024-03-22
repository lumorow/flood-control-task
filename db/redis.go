package db

import (
	"fmt"
	"github.com/go-redis/redis"
)

type Config struct {
	Host string
	Port string
}

func NewRedisDB(cfg Config) (*redis.Client, error) {
	r := redis.NewClient(&redis.Options{Addr: fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)})

	_, err := r.Ping().Result()
	if err != nil {
		return nil, err
	}

	return r, nil
}
