package dto

type KafkaEvent struct {
	Name string `json:"name"`
}

type KafkaMetaData struct {
	Sender    string `json:"sender"`
	SendingAt string `json:"sending_at"`
}

type DataType string

type KafkaBody[T any] struct {
	Type DataType `json:"type"`
	Data T        `json:"data"`
}

type KafkaMessage[T any] struct {
	Event    KafkaEvent    `json:"event"`
	MetaData KafkaMetaData `json:"meta_data"`
	Body     KafkaBody[T]  `json:"body"`
}
