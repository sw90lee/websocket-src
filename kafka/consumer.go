package kafka

import (
	"fmt"
	"github.com/IBM/sarama"
	"github.com/insoft-cloud/websocket-go/configuration"
	"github.com/insoft-cloud/websocket-go/server"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var cfg = configuration.InitConfig()

type ConsumerHandler struct {
	hub *server.Hub
}

// consumer로 사용시
func InitKafkaConsumer(hub *server.Hub) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	// Sarama Consumer 생성
	kfk, err := sarama.NewConsumer(cfg.Kafka.Broker, config)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := kfk.Close(); err != nil {
			panic(err)
		}
	}()

	// Kafka Topic에 연결
	partitions, err := kfk.Partitions(cfg.Kafka.Topic)
	if err != nil {
		panic(err)
	}

	// 각 Partition 별 Consumer 생성 및 Consume 시작
	var wg sync.WaitGroup
	wg.Add(len(partitions))

	for _, partition := range partitions {
		pc, err := kfk.ConsumePartition(cfg.Kafka.Topic, partition, sarama.OffsetNewest)
		if err != nil {
			fmt.Printf("Failed to start consumer for partition %d: %v\n", partition, err)
			return
		}

		go func(pc sarama.PartitionConsumer) {
			defer wg.Done()
			for msg := range pc.Messages() {
				hub.Broadcast <- msg.Value
				fmt.Printf("Partition: %d, Offset: %d, Key: %s, Value: %s\n",
					msg.Partition, msg.Offset, string(msg.Key), string(msg.Value))
			}
		}(pc)
	}

	// 시그널 처리를 통해 Consumer 정상 종료 처리
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sigchan

	// Consumer 종료를 기다림
	wg.Wait()
}
