package dto

type OutboxMessage struct {
	ID      int64
	Topic   string
	Payload []byte
}
