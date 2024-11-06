package conexoes

import (
	"fmt"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQConnection struct {
	conn    *amqp.Connection
	Channel *amqp.Channel
}

func ConectarRabbitMQ() (*RabbitMQConnection, error) {
	conn, err := amqp.Dial(fmt.Sprintf(
		"amqp://%s:%s@%s:%s/",
		os.Getenv("RABBITMQ_USER"), os.Getenv("RABBITMQ_PASSWORD"),
		os.Getenv("RABBITMQ_HOST"), os.Getenv("RABBITMQ_PORT"),
	))

	if err != nil {
		return nil, err
	}

	channel, err := conn.Channel()

	if err != nil {
		return nil, err
	}

	return &RabbitMQConnection{
		conn:    conn,
		Channel: channel,
	}, nil
}

func (conexao *RabbitMQConnection) Fechar() {
	conexao.Channel.Close()
	conexao.conn.Close()
}
