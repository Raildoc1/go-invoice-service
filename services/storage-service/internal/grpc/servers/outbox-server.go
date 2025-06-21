package servers

import (
	"context"
	pb "go-invoice-service/common/protocol/proto/messagescheduler"
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
}

func (o *OutboxServer) Get(ctx context.Context, request *pb.GetMessagesRequest) (*pb.GetMessagesResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (o *OutboxServer) Delete(ctx context.Context, request *pb.DeleteMessageRequest) (*pb.DeleteMessageResponse, error) {
	//TODO implement me
	panic("implement me")
}
