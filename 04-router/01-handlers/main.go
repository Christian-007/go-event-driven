package main

import (
	"context"
	"os"
	"strconv"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
)

const celsiusTopic = "temperature-celsius"
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

	pub, err := redisstream.NewPublisher(redisstream.PublisherConfig{
		Client: rdb,
	}, logger)
	if err != nil {
		panic(err)
	}

	router.AddHandler(
		"convert_celsius_to_fahrenheit_handler",
		celsiusTopic,
		sub,
		fahrenheitTopic,
		pub,
		convertCelsiusToFahrenheitHandler,
	)

	err = router.Run(context.Background())
	if err != nil {
		panic(err)
	}
}

func convertCelsiusToFahrenheitHandler(msg *message.Message) ([]*message.Message, error) {
	celsiusTemp := string(msg.Payload)
	fahrenheitTemp, err := celsiusToFahrenheit(celsiusTemp)
	if err != nil {
		return nil, err
	}

	newMsg := message.NewMessage(watermill.NewUUID(), []byte(fahrenheitTemp))
	return []*message.Message{newMsg},nil
}

func celsiusToFahrenheit(temperature string) (string, error) {
	celsius, err := strconv.Atoi(temperature)
	if err != nil {
		return "", err
	}

	return strconv.Itoa(celsius*9/5 + 32), nil
}
