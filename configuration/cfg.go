package configuration

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
	"strconv"
)

type Config struct {
	Kafka Kafka
}
type Kafka struct {
	Topic         string   `mapstructure:"topic"`
	Broker        []string `mapstructure:"broker"`
	Consumergroup string   `mapstructure:"consumergroup"`
	Partition     string   `mapstructure:"partition"`
	InitialOffset string   `mapstructure:"initialoffset"`
	Assignor      string   `mapstructure:"assignor"`
	Oldest        bool     `mapstructure:"oldest"`
}

func InitConfig() Config {
	// viper 설정
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")

	// viper overide
	viper.AutomaticEnv()

	// viper defaultSet 설정
	viper.SetDefault("kafka.topic", getEnv("Topic", ""))
	viper.SetDefault("kafka.broker", getEnv("broker", ""))
	viper.SetDefault("kafka.consumergroup", getEnv("Consumergroup", ""))
	viper.SetDefault("kafka.partition", getEnv("Partition", ""))
	viper.SetDefault("kafka.assignor", getEnv("Assignor", "roundrobin"))
	viper.SetDefault("kafka.oldest", getEnvAsBool("Oldest", false))

	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Printf("Config 파일 로드 중 에러: %v\n", err)
		}
		fmt.Println("Config 파일이 존재하지 않아 기본값을 사용합니다.")
	}

	var config Config
	err = viper.Unmarshal(&config)
	if err != nil {
		fmt.Println("config 매핑 에러")
	}

	fmt.Printf("Topic: %s\n", config.Kafka.Topic)
	fmt.Printf("Broker: %v\n", config.Kafka.Broker)
	fmt.Printf("ConsumerGroup: %v\n", config.Kafka.Consumergroup)

	return config
}

// env String 반환
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// env String 반환
func getEnvAsBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	valueBool, _ := strconv.ParseBool(value)
	return valueBool
}

// env int 반환
func getEnvAsInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	var intValue int
	_, err := fmt.Scan(value, &intValue)
	if err != nil {
		return defaultValue
	}
	return intValue
}
