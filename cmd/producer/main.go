package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type App struct {
	Channel *amqp.Channel
	Queue   amqp.Queue
}

func main() {
	//Подключение к RabbitMQ
	rmqURL := os.Getenv("RMQ_URL")
	if rmqURL == "" {
		rmqURL = "amqp://guest:guest@localhost:5672/"
	}

	conn, err := amqp.Dial(rmqURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"test_queue", // name
		false,        // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	app := &App{
		Channel: ch,
		Queue:   q,
	}

	//Настройка HTTP сервера
	http.HandleFunc("/send", app.sendMessageHandler)

	srv := &http.Server{
		Addr: ":8080",
	}

	// Запуск сервера в горутине
	go func() {
		log.Println("Producer listening on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe: %v", err)
		}
	}()

	// 3. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down producer...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
	log.Println("Producer shut down gracefully")
}

func (app *App) sendMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var msg struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, "Bad JSON", http.StatusBadRequest)
		return
	}

	// Публикуем сообщение, используя уже открытый канал
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := app.Channel.PublishWithContext(ctx,
		"",             // exchange
		app.Queue.Name, // routing key
		false,          // mandatory
		false,          // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(msg.Message),
		})

	if err != nil {
		http.Error(w, "Failed to publish message", http.StatusInternalServerError)
		log.Printf("Error publishing: %v", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Message sent to RMQ")
}
