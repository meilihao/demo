//  分发一个消息给多个消费者（consumers）。这种模式被称为“发布／订阅”
package main

import (
	"math/rand"
	"strconv"

	"time"

	"fmt"

	"github.com/streadway/amqp"
)

func CheckErr(err error) {
	if err != nil {
		panic(err)
	}
}

func randomString(l int) string {
	bytes := make([]byte, l)
	for i := 0; i < l; i++ {
		bytes[i] = byte(randInt(65, 90))
	}
	return string(bytes)
}

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@192.168.1.201:5672/chenz057")
	CheckErr(err)
	//defer conn.Close()

	ch, err := conn.Channel()
	CheckErr(err)
	//defer ch.Close()

	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when usused
		true,  // exclusive
		false, // noWait
		nil,   // arguments
	)
	CheckErr(err)

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	CheckErr(err)

	fmt.Println("queue name :", q.Name)

	for {
		corrId := randomString(32)

		n := int(time.Now().Unix()) % 10
		err = ch.Publish(
			"",          // exchange
			"rpc_queue", // routing key
			false,       //exchange根据自身类型和消息routeKey无法找到一个符合条件的queue，那么会调用channel.NotifyReturn方法将消息返还给生产者；当mandatory设为false时，出现上述情形broker会直接将消息扔掉。
			false,       //immediate在rabbitmq3.0中已删除.immediate=true且在for循环中会导致conn被关闭,原因未知.
			amqp.Publishing{
				ContentType:   "text/plain",
				CorrelationId: corrId,
				ReplyTo:       q.Name,
				Body:          []byte(strconv.Itoa(n)),
			})
		CheckErr(err)

		for d := range msgs {
			if corrId == d.CorrelationId {
				res, err := strconv.Atoi(string(d.Body))
				CheckErr(err)
				fmt.Printf("%d -> %d\n", n, res)
				break
			}
		}

		time.Sleep(time.Second * 2)
	}
}
