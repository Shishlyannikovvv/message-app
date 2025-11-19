package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	HandleFunchttp.("/send", sendMessageHandler)
	log.Println("Producer listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func sendMessageHandler(w http.ResponseWriter, r *http.Request) {
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

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		http.Error(w, "RMQ connect failed", http.StatusInternalServerError)
		log.Printf("Error: %v", err)
		return
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		http.Error(w, "RMQ channel failed", http.StatusInternalServerError)
		log.Printf("Error: %v", err)
		return
	}
	defer ch.Close()

	q, err := ch.QueueDeclare("test_queue", false, false, false, false, nil)
	if err != nil {
		http.Error(w, "Queue declare failed", http.StatusInternalServerError)
		log.Printf("Error: %v", err)
		return
	}

	err = ch.Publish("", q.Name, false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(msg.Message),
	})
	if err != nil {
		http.Error(w, "Publish failed", http.StatusInternalServerError)
		log.Printf("Error: %v", err)
		return
	}

	fmt.Fprintln(w, "Message sent to RMQ")
}