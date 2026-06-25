package kafka

type Config struct {
	Brokers              []string
	Auth                 bool // this is for cases where we need to authenticate the reader
	Username             string
	Password             string
	Topic, ConsumerGroup string // we might separate this config as only the reader needs a topic/consumerGroup at creation time.
}
