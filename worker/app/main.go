package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	_ "github.com/joho/godotenv/autoload"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/vitorbiten/maintenance/worker/app/controllers"
	"golang.org/x/sync/errgroup"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

var controllersMap map[string]func(delivery amqp.Delivery) = map[string]func(delivery amqp.Delivery){
	"notification": controllers.Notification,
}

func main() {
	var conn *amqp.Connection
	var err error
	for {
		conn, err = amqp.Dial(fmt.Sprintf(
			"amqp://%s:%s@%s:5672/",
			os.Getenv("RABBITMQ_USER"),
			os.Getenv("RABBITMQ_PASSWORD"),
			os.Getenv("RABBITMQ_HOST"),
		))
		if err == nil {
			break
		}
		log.Printf("Error connecting to rabbitmq: %s", err)
		log.Println("Retrying in 5 seconds...")
		time.Sleep(5 * time.Second)
	}
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"task_queue", // name
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.Qos(
		10,    // prefetch count
		0,     // prefetch size
		false, // global
	)
	failOnError(err, "Failed to set QoS")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	mainCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	closeNotifier := conn.NotifyClose(make(chan *amqp.Error))
	g, gCtx := errgroup.WithContext(mainCtx)

	var processingWg sync.WaitGroup

	g.Go(func() error {
		select {
		case <-closeNotifier:
			log.Println("RabbitMQ connection closed")
			stop()
			return nil
		case <-gCtx.Done():
			log.Println("Shutting down server...")
			return gCtx.Err()
		}
	})
	g.Go(func() error {
		for d := range msgs {
			select {
			case <-gCtx.Done():
				return gCtx.Err()
			default:
				incomingController := d.Headers["controller"].(string)
				failOnError(err, "Controller not found")
				processingWg.Add(1)
				go func(d amqp.Delivery) {
					defer processingWg.Done()
					controllersMap[incomingController](d)
				}(d)
			}
		}
		return nil
	})

	log.Printf(" [*] Waiting for messages...")
	<-gCtx.Done()
	log.Println("Awaiting final messages...")
	processingWg.Wait()
	conn.Close()
}
