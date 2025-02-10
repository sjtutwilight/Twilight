package ports

type Consumer interface {
	Consume() (*Message, error)
	Close()
}
