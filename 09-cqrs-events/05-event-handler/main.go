package main

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
)

type FollowRequestSent struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type EventsCounter interface {
	CountEvent() error
}

type FollowRequestSentHandler struct {
	counter EventsCounter
}

func (f FollowRequestSentHandler) Handle(ctx context.Context, event *FollowRequestSent) error {
	return f.counter.CountEvent()
}

func NewFollowRequestSentHandler(counter EventsCounter) cqrs.EventHandler {
	f := FollowRequestSentHandler{
		counter: counter,
	}

	return cqrs.NewEventHandler(
		"FollowRequestSentHandler",
		f.Handle,
	)
}
