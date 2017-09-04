package main

import (
	"fmt"
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
	//defer conn.Close()

	ch, err := conn.Channel()
	CheckErr(err)
	//defer ch.Close()

	err = ch.ExchangeDeclare(
		"logs_direct",
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	CheckErr(err)

	body := "publish_subscribe"

	r := make(chan amqp.Return)
	go func() {
		ch.NotifyReturn(r)

		for v := range r {
			fmt.Println("---", v)
		}
	}()

	t := time.NewTicker(time.Second)
	defer t.Stop()

	for {
		tmp := <-t.C

		fmt.Println(tmp.Unix())
		var _level string
		if tmp.Unix()%2 == 0 {
			_level = "orange"
		} else {
			_level = "green"
		}
		err = ch.Publish(
			"logs_direct", // exchange
			_level,        // routing key
			false,         //exchange根据自身类型和消息routeKey无法找到一个符合条件的queue，那么会调用channel.NotifyReturn方法将消息返还给生产者；当mandatory设为false时，出现上述情形broker会直接将消息扔掉。
			false,         //immediate在rabbitmq3.0中已删除.immediate=true且在for循环中会导致conn被关闭,原因未知.
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(body + ":" + _level),
			})
		CheckErr(err)
	}
}
