module validation-service

go 1.24.0

require (
	github.com/confluentinc/confluent-kafka-go/v2 v2.10.1
	go-invoice-service/common v0.0.0-00010101000000-000000000000
)

require github.com/google/uuid v1.6.0 // indirect

replace go-invoice-service/common => ./../../common
