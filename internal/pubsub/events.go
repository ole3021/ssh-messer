package pubsub

import "context"

// const (
// 	CreatedEvent EventType = "created"
// 	UpdatedEvent EventType = "updated"
// 	DeletedEvent EventType = "deleted"
// )

type Subscriber[E any, T any] interface {
	Subscribe(context.Context) <-chan Event[E, T]
}

type (
	// EventType identifies the type of event
	EventType string

	// Event represents an event in the lifecycle of a resource
	Event[E any, T any] struct {
		Type    E
		Payload T
	}

	Publisher[E any, T any] interface {
		Publish(E, T)
	}
)
