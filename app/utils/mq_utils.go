package utils

import (
	"BaseGoUni/core/utils"
	"encoding/json"
	"fmt"
	"github.com/wagslane/go-rabbitmq"
	"log"
	"strconv"
	"strings"
)

var MQ *rabbitmq.Conn

func InitRabbitMq() {
	var err error
	MQ, err = rabbitmq.NewConn(fmt.Sprintf("amqp://%s:%s@%s:%d/",
		utils.GlobalConfig.RabbitMq.Username, utils.GlobalConfig.RabbitMq.Password, utils.GlobalConfig.RabbitMq.Host, utils.GlobalConfig.RabbitMq.Port),
		rabbitmq.WithConnectionOptionsLogging)
	if err != nil {
		panic(err)
		return
	}
	defer MQ.Close()
	utils.Publisher, err = rabbitmq.NewPublisher(
		MQ,
		rabbitmq.WithPublisherOptionsLogging,
		rabbitmq.WithPublisherOptionsExchangeName(utils.GlobalConfig.RabbitMq.Exchange),
		rabbitmq.WithPublisherOptionsExchangeDeclare,
	)
	if err != nil {
		panic(err)
		return
	}
	defer utils.Publisher.Close()
	ListenNormal()
}

func ListenNormal() {
	consumer, err := rabbitmq.NewConsumer(MQ, utils.GlobalConfig.RabbitMq.Queue,
		rabbitmq.WithConsumerOptionsRoutingKey(utils.GlobalConfig.RabbitMq.Routing),
		rabbitmq.WithConsumerOptionsExchangeName(utils.GlobalConfig.RabbitMq.Exchange),
		rabbitmq.WithConsumerOptionsExchangeDeclare,
	)
	if err != nil {
		panic(err)
		return
	}
	defer consumer.Close()
	err = consumer.Run(MqReceiver)
	if err != nil {
		panic(err)
		return
	}
}

func MqReceiver(d rabbitmq.Delivery) rabbitmq.Action {
	log.Printf("consumed: %v", string(d.Body))
	var mqMessage utils.MQMessage
	_ = json.Unmarshal(d.Body, &mqMessage)
	switch mqMessage.MessageType {
	case utils.KeyCleanManageLog:
		utils.CleanManageLog(mqMessage.Data, mqMessage.DataMore)
		break
	case utils.KeyManageLogNotify:
		utils.NotifyManageLog(mqMessage.Data, mqMessage.DataMore)
		break
	case utils.KeyMqUserUpdate:
		datas := strings.Split(mqMessage.Data, "#")
		if len(datas) == 2 {
			userId, _ := strconv.ParseInt(datas[0], 10, 64)
			utils.FlushTempUser(datas[1], userId)
		}
		break
	}
	return rabbitmq.Ack
}
