package client

type EventBus interface {
	// Start receiving events. Events are distributed in the current consumer-group, so not every consumer receives all events.
	Subscribe(eventName, consumerGroup string, autoAck bool) (Subscription, error)

	// Stops all subscribes and closes all internal connections
	Close() error

	// Publish sends the given payload to all consumers subscribed to the given event.
	Publish(eventName string, payload interface{}) error

	// PublishAck(eventName string, payload interface{}) error
}

type Subscription interface {
	Events() <-chan Event

	Close() error
}

type Event interface {
	GetCorrelationId() string

	// Parse parses the message into the given object.
	Parse(object interface{}) error

	Ack() error
	Nack(requeue bool) error
	Reject(requeue bool) error
}
