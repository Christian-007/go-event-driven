package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
)

const fahrenheitTopic = "temperature-fahrenheit"

func main() {
	logger := watermill.NewStdLogger(false, false)

	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		panic(err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})

	sub, err := redisstream.NewSubscriber(redisstream.SubscriberConfig{
		Client: rdb,
	}, logger)
	if err != nil {
		panic(err)
	}

	router.AddNoPublisherHandler(
		"print_temperature_value",
		fahrenheitTopic,
		sub,
		func(msg *message.Message) error {
			value := string(msg.Payload)
			fmt.Printf("Temperature read: %v\n", value)

			return nil
		},
	)

	err = router.Run(context.Background())
	if err != nil {
		panic(err)
	}
}
