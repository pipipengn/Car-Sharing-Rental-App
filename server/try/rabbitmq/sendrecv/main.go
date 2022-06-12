package main

import (
	"fmt"
	"github.com/streadway/amqp"
	"time"
)

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		panic(err)
	}

	ch, err := conn.Channel()
	q, err := ch.QueueDeclare(
		"go_q1",
		true,  // durable
		false, // autoDelete
		false, // exlusive
		false, // noWait
		nil,   // args
	)
	if err != nil {
		panic(err)
	}

	go consume("c1", conn, q.Name)
	go consume("c2", conn, q.Name)

	i := 0
	for {
		i++
		if err := ch.Publish("", q.Name, false, false, amqp.Publishing{Body: []byte(fmt.Sprintf("message %d", i))}); err != nil {
			fmt.Println(err.Error())
		}
		time.Sleep(time.Millisecond * 200)
	}
}

func consume(consumer string, conn *amqp.Connection, q string) {
	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}

	msgs, err := ch.Consume(
		q,
		consumer, // consumer
		true,     // autoAck
		false,    // exclusive
		false,    // noLocal
		false,    // noWait
		nil,      // args
	)
	if err != nil {
		panic(err)
	}

	for msg := range msgs {
		fmt.Printf("%s:%s\n", consumer, msg.Body)
	}
}
