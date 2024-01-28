package controllers

import (
	"encoding/json"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func Notification(delivery amqp.Delivery) {
	responseMap := make(map[string]interface{})
	err := json.Unmarshal(delivery.Body, &responseMap)
	if err != nil {
		err := delivery.Nack(false, false)
		if err != nil {
			log.Panicf("%s", err)
		}
	}
	log.Printf("The tech %s performed the task %s on date %s\n",
		responseMap["nickname"],
		responseMap["task_id"],
		responseMap["task_date"],
	)
	time.Sleep(1 * time.Second)
	log.Printf("Acknowledging message %d\n", delivery.DeliveryTag)
	err = delivery.Ack(false)
	if err != nil {
		log.Panicf("%s", err)
	}
}
