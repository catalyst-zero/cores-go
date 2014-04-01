package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	cores "github.com/catalyst-zero/cores-go"
	"time"
)

func newHash(data string) string {
	hash := sha256.New()
	hash.Write([]byte(data))

	return hex.EncodeToString(hash.Sum(nil))
}

type TransactionFinishedEvent struct {
	Id string `json:"id"`
}

func NewTransactionFinishedEvent() TransactionFinishedEvent {
	return TransactionFinishedEvent{
		Id: newHash(fmt.Sprintf("%d", time.Now())),
	}
}

func main() {
	eventbus, err := cores.NewAmqpClient("amqp://:5672", "example-producer")
	if err != nil {
		panic(err)
	}

	producer, err := eventbus.CreateProducer(cores.ProducerOptions{EventName: "transaction-finished"})

	for {
		time.Sleep(time.Second)

		for i := 0; i < 1000; i++ {
			payload := NewTransactionFinishedEvent()
			fmt.Printf("Sending %s\n", payload.Id)
			producer.Send(payload)
		}
	}
}
