package services

import (
	"context"
	"fmt"
	pb "go-invoice-service/common/protocol/proto/messagescheduler"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/durationpb"
	"message-sheduler-service/internal/dto"
	"time"

	"google.golang.org/grpc"
)

type StorageConfig struct {
	ServerAddress string
}

type Storage struct {
	conn                *grpc.ClientConn
	outboxStorageClient pb.OutboxStorageClient
}

func NewStorage(cfg StorageConfig) (*Storage, error) {
	options := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.NewClient(cfg.ServerAddress, options)
	if err != nil {
		return nil, err
	}
	outboxStorageClient := pb.NewOutboxStorageClient(conn)
	return &Storage{
		conn:                conn,
		outboxStorageClient: outboxStorageClient,
	}, nil
}

func (s *Storage) Close() error {
	err := s.conn.Close()
	if err != nil {
		return fmt.Errorf("failed to close gRPC connection: %w", err)
	}
	return nil
}

func (s *Storage) GetOutboxMessages(
	ctx context.Context,
	maxCount int32,
	retryIn time.Duration,
) ([]dto.OutboxMessage, error) {
	req := &pb.GetMessagesRequest{
		MaxCount: &maxCount,
		RetryAfter: &durationpb.Duration{
			Seconds: int64(retryIn.Seconds()),
		},
	}
	resp, err := s.outboxStorageClient.Get(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get outbox messages: %w", err)
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("failed to get outbox messages: %s", *resp.Error)
	}
	msgsCount := len(resp.OutboxMessages)
	res := make([]dto.OutboxMessage, msgsCount)
	for i, msg := range resp.OutboxMessages {
		res[i] = dto.OutboxMessage{
			ID:      msg.GetId(),
			Topic:   msg.GetTopic(),
			Payload: msg.GetPayload(),
		}
	}
	return res, nil
}

func (s *Storage) DeleteOutboxMessage(ctx context.Context, id int64) error {
	req := &pb.DeleteMessageRequest{
		Id: &id,
	}
	resp, err := s.outboxStorageClient.Delete(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to delete outbox message: %w", err)
	}
	if resp.Error != nil {
		return fmt.Errorf("failed to delete outbox message: %s", *resp.Error)
	}
	return nil
}
