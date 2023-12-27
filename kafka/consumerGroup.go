package kafka

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/gorilla/websocket"
	"github.com/insoft-cloud/websocket-go/server"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

var (
	brokers  = cfg.Kafka.Broker
	group    = cfg.Kafka.Consumergroup
	topics   = cfg.Kafka.Topic
	assignor = cfg.Kafka.Assignor
	oldest   = cfg.Kafka.Oldest
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type ConsumerGroupHandler struct {
	ready   chan bool
	wsConn  *websocket.Conn
	clients *server.Client
	hub     *server.Hub
}

func InitConsumerGroup(hub *server.Hub) {
	keepRunning := true
	log.Println("Starting a new Sarama consumer")

	config := sarama.NewConfig()
	switch assignor {
	case "sticky":
		config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategySticky()}
	case "roundrobin":
		config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	case "range":
		config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRange()}
	default:
		log.Panicf("Unrecognized consumer group partition assignor: %s", assignor)
	}

	if oldest {
		config.Consumer.Offsets.Initial = sarama.OffsetOldest
	}

	kfkConsumer := ConsumerGroupHandler{
		ready: make(chan bool),
		hub:   hub,
	}
	ctx, cancel := context.WithCancel(context.Background())
	client, err := sarama.NewConsumerGroup(brokers, group, config)
	if err != nil {
		log.Panicf("Error creating consumer group client: %v", err)
	}

	// 소비 일시중기 기능
	consumptionIsPaused := false
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		for {
			// consumer는 무한루프에서 호출 되어야함
			// server-side에서 rebalance하면 소비자 세션은 new claim을 재생성
			if err := client.Consume(
				ctx,
				strings.Split(topics, ","), // string을 ,분리하여 []string배열로 만듬
				&kfkConsumer,
			); err != nil {
				log.Panicf("Error from consumer: %v", err)
			}

			//context 취소를 확인하여 consumer가 중지해야함을 알림
			if ctx.Err() != nil {
				return
			}

			kfkConsumer.ready = make(chan bool)
		}
	}()

	<-kfkConsumer.ready // Consumer 설정될 때까지 대기
	log.Println("Sarama consumer up and running!...")
	sigusr1 := make(chan os.Signal, 1)
	signal.Notify(sigusr1, os.Interrupt)

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	for keepRunning {
		select {
		case <-ctx.Done():
			log.Println("terminating: context cancelled")
			keepRunning = false
		case <-sigterm:
			log.Println("terminating: via signal")
			keepRunning = false
		case <-sigusr1:
			toggleConsumptionFlow(client, &consumptionIsPaused)
		}
	}

	cancel()
	wg.Wait()
	if err = client.Close(); err != nil {
		log.Panicf("Error closing client: %v", err)
	}
}

func toggleConsumptionFlow(client sarama.ConsumerGroup, isPaused *bool) {
	if *isPaused {
		client.ResumeAll()
		log.Println("Resuming consumption")
	} else {
		client.PauseAll()
		log.Println("Pausing consumption")
	}

	*isPaused = !*isPaused
}

// Consumer Group 생성
func (h *ConsumerGroupHandler) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h *ConsumerGroupHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				log.Printf("message channel was closed")
				return nil
			}
			log.Printf("Message claimed: value = %s, timestamp = %v, topic = %s", string(message.Value), message.Timestamp, message.Topic)
			session.MarkMessage(message, "")

			// hub Broadcast 로 메세지 보내기
			go func() {
				h.hub.Broadcast <- message.Value
			}()
		case <-session.Context().Done():
			return nil
		}
	}
}
