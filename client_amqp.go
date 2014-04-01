package client

func NewAmqpClient(addr string) (Client, error) {
	return new(amqpClient), nil
}

type amqpClient struct {
}

// Returns the name of this consumer group, e.g. "orlok" or "hades"
func (client *amqpClient) GetConsumerGroup() string {
	return "unknown"
}

// Publish an event to all consumer groups
func (client *amqpClient) Broadcast(eventName string, payload interface{}) error {
	return nil
}

// Start receiving events. Events are distributed in the current consumer-group, so not every consumer receives all events.
func (client *amqpClient) Subscribe(eventName string, handler Handler) error {
	return nil
}

// Stops all subscribes and closes all internal connections
func (client *amqpClient) Close() error {
	return nil
}
