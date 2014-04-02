package client

type EventBus interface {
	// Returns the name of this consumer group, e.g. "orlok" or "hades"
	GetConsumerGroup() string

	// Start receiving events. Events are distributed in the current consumer-group, so not every consumer receives all events.
	Subscribe(eventName string, autoAck bool, handler Handler) error

	// Stops all subscribes and closes all internal connections
	Close() error

	Publish(eventName string, payload interface{}) error
	// PublishAck(eventName string, payload interface{}) error
}

type Handler interface {
	Handle(event Event)
}

type Event interface {
	GetCorrelationId() string

	// Parse parses the message into the given object.
	Parse(object interface{}) error

	Ack() error
	Nack(requeue bool) error
	Reject(requeue bool) error
}
