package main

import (
	"bytes"
	"log"
	"time"

	"github.com/streadway/amqp"
)

func CheckErr(err error) {
	if err != nil {
		panic(err)
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
		"task_queue",
		true,
		false,
		false,
		false,
		nil,
	)
	CheckErr(err)

	err = ch.Qos(
		1,     // prefetch count,一次处理消息的最大数量,autoack=true,会忽略该参数
		0,     // prefetch size,可接收消息的大小,0为没限制
		false, // global,是否针对整个conn进行调整,false为仅当前channel
	)
	CheckErr(err)

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer的唯一标识
		false,  // auto-ack,消费端不用发送Ack
		false,  // exclusive
		false,  // no-local,rabbitmq不支持该参数
		false,  // no-wait
		nil,    // args
	)
	CheckErr(err)

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			dot_count := bytes.Count(d.Body, []byte("."))
			time.Sleep(time.Duration(dot_count) * time.Second)
			log.Printf("Done")
			d.Ack(false) // multiple:fasle,只确认该delivery;true,DeliveryTag小于等于d.DeliveryTag且未确认的都会被确认.
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
