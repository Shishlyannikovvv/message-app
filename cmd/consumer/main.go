package main

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("RMQ connect failed: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("RMQ channel failed: %v", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare("test_queue", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("Queue declare failed: %v", err)
	}

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Consume failed: %v", err)
	}

	log.Println("Consumer waiting for messages... Press CTRL+C to exit.")

	for d := range msgs {
		log.Printf("Received: %s", d.Body)
	}
}
