package servers

import (
	"context"
	"fmt"
	pb "go-invoice-service/common/protocol/proto/messagescheduler"
	"go-invoice-service/common/protocol/proto/types"
	"google.golang.org/protobuf/types/known/emptypb"
	"storage-service/internal/dto"
	"time"
)

var _ pb.OutboxStorageServer = (*OutboxServer)(nil)

type OutboxService interface {
	Get(ctx context.Context, maxCount int32, retryAfter time.Duration) ([]dto.OutboxMessage, error)
	Delete(ctx context.Context, id int64) error
}

type OutboxServer struct {
	pb.UnimplementedOutboxStorageServer
	outboxService OutboxService
}

func NewOutboxServer(outboxService OutboxService) *OutboxServer {
	return &OutboxServer{
		outboxService: outboxService,
	}
}

func (o *OutboxServer) Get(ctx context.Context, request *pb.GetMessagesRequest) (*pb.GetMessagesResponse, error) {
	messages, err := o.outboxService.Get(ctx, request.GetMaxCount(), request.GetRetryAfter().AsDuration())
	if err != nil {
		return nil, fmt.Errorf("failed to get outbox messages: %w", err)
	}
	response := &pb.GetMessagesResponse{
		OutboxMessages: convertMessages(messages),
	}
	return response, nil
}

func convertMessages(messages []dto.OutboxMessage) []*types.OutboxMessage {
	res := make([]*types.OutboxMessage, len(messages))

	for i, message := range messages {
		res[i] = convertMessage(message)
	}

	return res
}

func convertMessage(message dto.OutboxMessage) *types.OutboxMessage {
	topicString := string(message.Topic)
	return &types.OutboxMessage{
		Id:      &message.ID,
		Topic:   &topicString,
		Payload: message.Payload,
	}
}

func (o *OutboxServer) Delete(ctx context.Context, request *pb.DeleteMessageRequest) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}
