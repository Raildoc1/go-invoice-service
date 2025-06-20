package kafka

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go-invoice-service/common/pkg/logging"
	common "go-invoice-service/common/protocol/kafka"
)

func Setup(ctx context.Context, address string, logger *logging.ZapLogger) error {
	for _, t := range common.Topics {
		err := ensureTopic(
			ctx,
			address,
			t.Topic,
			t.PartitionsCount,
			t.ReplicationFactor,
		)
		if err != nil {
			return fmt.Errorf("ensure topic %s failed: %w", t.Topic, err)
		}
		logger.InfoCtx(ctx, fmt.Sprintf("Topic %s ensured", t.Topic))
	}
	return nil
}

func ensureTopic(
	ctx context.Context,
	address string,
	topic common.Topic,
	partitionsCount int,
	replicationFactor int,
) error {
	admin, err := kafka.NewAdminClient(&kafka.ConfigMap{"bootstrap.servers": address})
	if err != nil {
		return err
	}
	defer admin.Close()

	results, err := admin.CreateTopics(
		ctx,
		[]kafka.TopicSpecification{{
			Topic:             string(topic),
			NumPartitions:     partitionsCount,
			ReplicationFactor: replicationFactor,
		}},
		kafka.SetAdminOperationTimeout(5e9),
	)
	if err != nil {
		return err
	}

	for _, res := range results {
		if res.Error.Code() != kafka.ErrNoError && res.Error.Code() != kafka.ErrTopicAlreadyExists {
			return fmt.Errorf("failed to create topic %s: %v", topic, res.Error)
		}
	}

	return nil
}
