package rabbitmq

import (
	"fmt"
	"log"
	"os"

	"github.com/streadway/amqp"
)

// type Rabbit {

// }
type RabbitChannel struct {
	*amqp.Channel
}

func GetRabbitConn() *RabbitChannel {
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%s", os.Getenv("RABBIT_USER"), os.Getenv("RABBIT_PASS"), os.Getenv("RABBIT_HOST"), os.Getenv("RABBIT_PORT")))
	if err != nil {
		log.Println("Unable to connect to rabbitmq ", err.Error())
		os.Exit(1)
	}
	rabChannel, err := conn.Channel()
	if err != nil {
		log.Println("Unable to create channel to rabbitmq ", err.Error())
		os.Exit(1)
	}
	return &RabbitChannel{rabChannel}
}
