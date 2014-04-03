package main

import (
	"fmt"
	cores "github.com/catalyst-zero/cores-go"
	"time"
)

var _ = time.Now

type TransactionFinishedEvent struct {
	Id string `json:"id"`
}

var counter int = 0

func DoWork(transaction TransactionFinishedEvent) error {
	// fmt.Printf("> %s\n", transaction.Id)
	counter++

	if counter > 3 {
		counter = 0
		return fmt.Errorf("Something went wrong ...")
	} else {
		return nil
	}
}

func Handle(event cores.Event) {
	var transaction TransactionFinishedEvent

	if err := event.Parse(&transaction); err != nil {
		fmt.Printf("ERROR: Event#Parse() failed: %v\n", err)

		if err := event.Reject(false); err != nil {
			fmt.Printf("ERROR: Reject failed: %v\n", err)
			return
		}
		fmt.Printf("ERROR: Event#Parse() failed: %v\n", err)
		return
	}

	if err := DoWork(transaction); err != nil {
		if err := event.Nack(true); err != nil {
			fmt.Printf("ERROR: Nack() failed: %v\n", err)
		}
	} else {
		if err := event.Ack(); err != nil {
			fmt.Printf("ERROR: Ack() failed: %v\n", err)
		}
	}
}

func main() {
	eventbus, err := cores.NewAmqpEventBus("amqp://:5672")
	if err != nil {
		panic(err)
	}
	defer eventbus.Close()

	subscriber, err := eventbus.Subscribe("transaction-finished", "example-consumer", false)
	if err != nil {
		panic(err)
	}
	defer subscriber.Close()

	for event := range subscriber.Events() {
		Handle(event)
	}
}
