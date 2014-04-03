package main

import (
	"fmt"
	amqp "github.com/streadway/amqp"
)

func main() {
	connection, err := amqp.Dial("amqp://:5672")
	if err != nil {
		panic(err)
	}

	channel, err := connection.Channel()
	if err != nil {
		panic(err)
	}

	errorChan := make(chan *amqp.Error)

	channel.NotifyClose(errorChan)

	go func() {
		for err := range errorChan {
			fmt.Printf("ERROR on channel: %v\n", err)
		}
	}()

	autoAck := false
	consumerGroup := "statistics"
	eventName := "transaction-finished"
	consumerTag := "" // fmt.Sprintf("cores-go#%s.%s", consumerGroup, eventName)
	queueName := fmt.Sprintf("%s.%s", consumerGroup, eventName)

	// In this example I assume the queue already exists and has many messages in it
	deliveryChan, err := channel.Consume(queueName, consumerTag, autoAck, false, false, false, nil)
	if err != nil {
		panic(err)
	}

	// The separate goroutine is not necessary, but mimicks my use-case
	go func() {
		counter := 0
		for msg := range deliveryChan {
			counter++
			if counter > 3 {
				counter = 0
				msg.Reject(false)
			} else {
				msg.Ack(false)
			}
		}
	}()

	// Block
	select {}
}
