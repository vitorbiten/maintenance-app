package adapters

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

var PublishMessages = func(messages []map[string]interface{}, controller string) error {
	conn, err := amqp.Dial(fmt.Sprintf(
		"amqp://%s:%s@%s:5672/",
		os.Getenv("RABBITMQ_USER"),
		os.Getenv("RABBITMQ_PASSWORD"),
		os.Getenv("RABBITMQ_HOST"),
	))
	if err != nil {
		return errors.New("failed to connect to RabbitMQ")
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		return errors.New("failed to open a channel")
	}
	defer ch.Close()
	q, err := ch.QueueDeclare(
		"task_queue", // name
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return errors.New("failed to declare a queue")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, message := range messages {
		jsonMessage, err := json.Marshal(message)
		if err != nil {
			return errors.New("failed to encode a message")
		}
		err = ch.PublishWithContext(ctx,
			"",     // exchange
			q.Name, // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				DeliveryMode: 2,
				ContentType:  "text/plain",
				Body:         []byte(jsonMessage),
				Headers: map[string]interface{}{
					"controller": controller,
				},
			})
		if err != nil {
			return errors.New("failed to publish a message")
		}
		log.Printf(" [x] Sent %s\n", jsonMessage)
	}
	return nil
}
