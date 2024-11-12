package main

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/message"
)

type AlarmClient interface {
	StartAlarm() error
	StopAlarm() error
}

func ConsumeMessages(sub message.Subscriber, alarmClient AlarmClient) {
	messages, err := sub.Subscribe(context.Background(), "smoke_sensor")
	if err != nil {
		panic(err)
	}

	for msg := range messages {
		smokeSensor := string(msg.Payload)
		if (smokeSensor == "0") {
			err = alarmClient.StopAlarm()
			if err != nil {
				msg.Nack()
				continue
			}

			msg.Ack()
			continue
		}

		err = alarmClient.StartAlarm()
		if err != nil {
			msg.Nack()
			continue
		}

		msg.Ack()
	}
}
