package main

import (
	"context"
	"github.com/go-redis/redis"
	"log"
	"strconv"
	"sync"
	"task/db"
)

type FloodControl interface {
	Check(ctx context.Context, userID int64) (bool, error)
	ResetCounter(ctx context.Context, db *redis.Client, userID int64) error
}

type FloodControlImpl struct {
	cfg FloodControlConfig
	mu  sync.Mutex
	db  *redis.Client
}

func NewFloodControl(db *redis.Client, cfg FloodControlConfig) FloodControl {
	return &FloodControlImpl{
		cfg: cfg,
		db:  db,
	}
}

func main() {
	ctx := context.Background()

	redisDB, err := db.NewRedisDB(db.Config{Host: "localhost", Port: "6379"})
	if err != nil {
		log.Fatal(err)
	}

	userId := int64(1)
	msg := []Message{
		{msg: "HELLO EVERYONE!"},
		{msg: "HELLO EVERYONE!"},
		{msg: "HELLO EVERYONE!"},
		{msg: "HELLO EVERYONE!"},
		{msg: "HELLO EVERYONE!"},
		{msg: "HELLO EVERYONE!"},
		{msg: "HELLO EVERYONE!"},
		{msg: "HELLO EVERYONE!"},
		{msg: "HELLO EVERYONE!"},
		{msg: "HELLO EVERYONE!"},
		{msg: "HELLO EVERYONE!"},
		{msg: "HELLO EVERYONE!"},
		{msg: "HELLO EVERYONE!"},
		{msg: "HELLO EVERYONE!"},
		{msg: "HELLO EVERYONE!"},
		{msg: "HELLO EVERYONE!"},
		{msg: "HELLO EVERYONE!"},
		{msg: "HELLO EVERYONE!"},
		{msg: "HELLO EVERYONE!"},
		{msg: "HELLO EVERYONE!"},
		{msg: "HELLO EVERYONE!"},
		{msg: "HELLO EVERYONE!"},
		{msg: "HELLO EVERYONE!"},
		{msg: "HELLO EVERYONE!"},
	}

	makeBatchCheckFlood(ctx, redisDB, userId, msg)
}

func makeBatchCheckFlood(ctx context.Context, redisDB *redis.Client, userID int64, msgs []Message) {
	cfg := FloodControlConfig{MaxMesseages: 15, Interval: 2}
	fl := NewFloodControl(redisDB, cfg)
	rateLimiter := NewRateLimiter(cfg.Interval)
	outputCh := make(chan string, len(msgs))
	done := make(chan struct{})

	defer func() {
		close(outputCh)
		close(done)
		err := fl.ResetCounter(ctx, redisDB, userID)
		if err != nil {
			log.Fatal(err)
		}
	}()

	for _, m := range msgs {
		if rateLimiter.allow() {
			m := m
			go func() {
				flood, err := fl.Check(ctx, userID)
				if err != nil {
					log.Print(err)
				}

				if flood {
					log.Printf("user with ID: %d is a flooder", userID)

					// Хочу выйти полностью
					done <- struct{}{}
					return
				}
				outputCh <- m.msg
			}()
		} else {
			err := fl.ResetCounter(ctx, redisDB, userID)
			if err != nil {
				log.Println(err)
			}
		}
	}

	select {
	case data := <-outputCh:
		sendMessage(userID, data)
	case <-done:
		return
	}
}

func (f *FloodControlImpl) Check(ctx context.Context, userID int64) (bool, error) {
	id := strconv.FormatInt(userID, 10)

	f.mu.Lock()
	defer f.mu.Unlock()

	data, err := f.db.Get(id).Result()
	if err != nil && err != redis.Nil {
		return false, err
	}

	var count int

	if data != "" {
		count, err = strconv.Atoi(data)
		if err != nil {
			return false, err
		}
	}

	if count >= f.cfg.MaxMesseages {
		return true, nil
	}

	err = f.db.Incr(id).Err()
	if err != nil {
		return false, err
	}

	return false, nil
}

func (f *FloodControlImpl) ResetCounter(ctx context.Context, db *redis.Client, userID int64) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	id := strconv.FormatInt(userID, 10)
	err := db.Set(id, "0", 0).Err()
	if err != nil {
		return err
	}

	return nil
}

func sendMessage(userID int64, message string) {
	log.Printf("%d send message: %s", userID, message)
}
