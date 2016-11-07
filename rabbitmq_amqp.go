package main

/**
Auth: ghostwwl
Email: ghostwwl@gmail.com
Note: 测试go的amqp库呢
**/

import (
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"
)

const (
	_MESSAGE_SERVER        = "192.168.3.127"
	_MESSAGE_USER          = "ghostwwl"
	_MESSAGE_PWD           = "123465"
	_MESSAGE_PORT          = 5672
	_MESSAGE_VHOST         = "/artxun"
	_MESSAGE_TASK_EXCHANGE = "ARTXUN_WALLETS_TASK"
	exchangeType           = "direct"
)

const (
	formatTime     = "15:04:05"
	formatDate     = "2006-01-02"
	formatDateTime = "2006-01-02 15:04:05"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	msg_send()
	//msg_recv()
}

func msg_publish(amqpURI, exchange, exchangeType, routingKey, msg_body string, reliable bool) error {
	log.Printf("Rabbitmq Server:%q", amqpURI)
	connection, err := amqp.Dial(amqpURI)
	if err != nil {
		return fmt.Errorf("Dial: %s", err)
	}
	defer connection.Close()

	log.Printf("获取一个 Channel 信道")
	channel, err := connection.Channel()
	if err != nil {
		return fmt.Errorf("开启 Channel 失败: %s", err)
	}
	defer channel.Close()

	log.Printf("交换机类型: %q 交换机 %q", exchangeType, exchange)
	if err := channel.ExchangeDeclare(
		exchange,     // name
		exchangeType, // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // noWait
		nil,          // arguments
	); err != nil {
		return fmt.Errorf("定义交换机失败: %s", err)
	}

	if reliable {
		log.Printf("开启消息发送结果回执模式")
		//log.Printf("enabling publishing confirms.")
		if err := channel.Confirm(false); err != nil {
			//return fmt.Errorf("Channel could not be put into confirm mode: %s", err)
			return fmt.Errorf("Channel 无法开启消息确认模式: %s", err)
		}

		confirms := channel.NotifyPublish(make(chan amqp.Confirmation, 1))

		defer confirmMsgPub(confirms)
	}

	msg_obj := amqp.Publishing{
		Headers:         amqp.Table{},
		ContentType:     "application/json",
		ContentEncoding: "utf-8",
		Body:            []byte(msg_body),
		DeliveryMode:    amqp.Persistent, // 1=Transient 非持久, 2=persistent 持久消息
		Timestamp:       time.Time(time.Now()),
		Type:            "test",
		//Priority:        0,              // 0-9
		//UserId:       0,
	}

	log.Printf("发布消息大小: %dB 消息正文: %q", len(msg_body), msg_body)
	if err = channel.Publish(
		exchange,   // publish to an exchange
		routingKey, // routing to 0 or more queues
		false,      // mandatory
		false,      // immediate
		msg_obj,
	); err != nil {
		return fmt.Errorf("Exchange Publish: %s", err)
	}

	return nil
}

func msg_send() {
	var amqp_server string = fmt.Sprintf("amqp://%s:%s@%s:%d/%s", _MESSAGE_USER, _MESSAGE_PWD, _MESSAGE_SERVER, _MESSAGE_PORT, _MESSAGE_VHOST)
	msg_body := `{"amount":"25.00","wechat_orderid":"4009312001201610197148140819"}`
	send_flag := msg_publish(amqp_server, _MESSAGE_TASK_EXCHANGE, "direct", "log", msg_body, true)
	log.Println(send_flag)
}

func confirmMsgPub(confirms <-chan amqp.Confirmation) {
	//log.Printf("waiting for confirmation of one publishing\n")
	log.Printf("等待发送结果回执确认......\n")

	if confirmed := <-confirms; confirmed.Ack {
		//log.Printf("confirmed delivery with delivery tag: %d\n", confirmed.DeliveryTag)
		log.Printf("消息发送成功回执标签: %d\n", confirmed.DeliveryTag)
	} else {
		//log.Printf("failed delivery of delivery tag: %d\n", confirmed.DeliveryTag)
		log.Printf("消息发送失败回执标签: %d\n", confirmed.DeliveryTag)
	}
}
