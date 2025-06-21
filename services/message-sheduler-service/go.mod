module message-sheduler-service

go 1.24.0

require (
	github.com/confluentinc/confluent-kafka-go/v2 v2.10.1
	go-invoice-service/common v0.0.0-00010101000000-000000000000
	go.uber.org/zap v1.27.0
	golang.org/x/sync v0.12.0
	google.golang.org/grpc v1.73.0
	google.golang.org/protobuf v1.36.6
)

require (
	github.com/google/uuid v1.6.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250324211829-b45e905df463 // indirect
)

replace go-invoice-service/common => ./../../common
