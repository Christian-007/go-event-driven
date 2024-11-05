package main

import (
	"log"
	"time"
)

type User struct {
	Email string
}

type UserRepository interface {
	CreateUserAccount(u User) error
}

type NotificationsClient interface {
	SendNotification(u User) error
}

type NewsletterClient interface {
	AddToNewsletter(u User) error
}

type Handler struct {
	repository          UserRepository
	newsletterClient    NewsletterClient
	notificationsClient NotificationsClient
}

func NewHandler(
	repository UserRepository,
	newsletterClient NewsletterClient,
	notificationsClient NotificationsClient,
) Handler {
	return Handler{
		repository:          repository,
		newsletterClient:    newsletterClient,
		notificationsClient: notificationsClient,
	}
}

func (h Handler) SignUp(u User) error {
	if err := h.repository.CreateUserAccount(u); err != nil {
		return err
	}

	go retry(func() error {
		return h.newsletterClient.AddToNewsletter(u)
	}, "failed to add user to the newsletter")

	go retry(func() error {
		return h.notificationsClient.SendNotification(u)
	}, "failed to send the notification to the user")

	return nil
}

func retry(f func() error, errorMessage string) {
	for {
		if err := f(); err != nil {
			log.Printf(errorMessage + ": %v", err)
			time.Sleep(1 * time.Second)
			continue
		}
		break
	}
}
