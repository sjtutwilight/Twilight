package ports

type Producer interface {
	Produce(topic string, key, value []byte) error
}
