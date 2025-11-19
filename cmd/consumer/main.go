package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	rmqURL := os.Getenv("RMQ_URL")
	if rmqURL == "" {
		rmqURL = "amqp://guest:guest@localhost:5672/"
	}

	// Подключение
	conn, err := amqp.Dial(rmqURL)
	if err != nil {
		log.Fatalf("RMQ connect failed: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("RMQ channel failed: %v", err)
	}
	defer ch.Close()

	// Объявление очереди
	q, err := ch.QueueDeclare(
		"test_queue", // name
		false,        // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		log.Fatalf("Queue declare failed: %v", err)
	}

	consumerTag := "simple-consumer"
	msgs, err := ch.Consume(
		q.Name,      // queue
		consumerTag, // consumer
		true,        // auto-ack
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)
	if err != nil {
		log.Fatalf("Consume failed: %v", err)
	}

	log.Println("Consumer waiting for messages... Press CTRL+C to exit.")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for d := range msgs {
			log.Printf("Received: %s", d.Body)
		}
	}()

	<-quit
	log.Println("Shutting down consumer...")

	if err := ch.Cancel(consumerTag, false); err != nil {
		log.Printf("Consumer cancel failed: %v", err)
	}

	time.Sleep(1 * time.Second)
	log.Println("Consumer shut down gracefully")
}
