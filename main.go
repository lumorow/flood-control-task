package main

import (
	"context"
	"fmt"
	"github.com/go-redis/redis"
	"task/flood_control"
)

func main() {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	fc := flood_control.NewFloodControl(redisClient, 60, 10)

	ok, err := fc.Check(context.Background(), 1)
	if err != nil {
		fmt.Println(err)
	} else if !ok {
		fmt.Println("user is a flooder")
	} else {
		fmt.Println("user is a OK!")
	}
}
