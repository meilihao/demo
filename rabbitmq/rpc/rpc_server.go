package main

import (
	"log"
	"strconv"

	"github.com/streadway/amqp"
)

func CheckErr(err error) {
	if err != nil {
		panic(err)
	}
}

func fib(n int) int {
	if n == 0 {
		return 0
	} else if n == 1 {
		return 1
	} else {
		return fib(n-1) + fib(n-2)
	}
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@192.168.1.201:5672/chenz057")
	CheckErr(err)
	defer conn.Close()

	ch, err := conn.Channel()
	CheckErr(err)
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"rpc_queue", // name
		false,       // durable
		false,       // delete when unused
		true,        // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	CheckErr(err)

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	CheckErr(err)

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	CheckErr(err)

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			n, err := strconv.Atoi(string(d.Body))
			CheckErr(err)

			response := fib(n)
			log.Printf(" [.] fib(%d) -> %d", n, response)

			err = ch.Publish(
				"",        // exchange
				d.ReplyTo, // 用来命名回调队列
				false,     // mandatory
				false,     // immediate
				amqp.Publishing{
					ContentType:   "text/plain",
					CorrelationId: d.CorrelationId, //用来将RPC的响应和请求关联起来
					Body:          []byte(strconv.Itoa(response)),
				})
			CheckErr(err)

			d.Ack(false)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
