package controllers

import (
	"context"
	"message-sheduler-service/internal/dto"
	"sync"
	"time"
)

type StorageService interface {
	GetOutboxMessages(ctx context.Context, maxCount int, retryIn time.Duration) ([]dto.OutboxMessage, error)
	DeleteOutboxMessage(ctx context.Context, id int64) error
}

type KafkaProducer interface {
	SendMessage(ctx context.Context, topic string, payload []byte) error
}

type OutboxDispatcher struct {
	cfg            Config
	storageService StorageService
	kafkaProducer  KafkaProducer
}

type Config struct {
	NumWorkers       int
	RetryIn          time.Duration
	DispatchInterval time.Duration
}

func NewOutboxDispatcher(cfg Config, storageService StorageService, kafkaProducer KafkaProducer) *OutboxDispatcher {
	return &OutboxDispatcher{
		cfg:            cfg,
		storageService: storageService,
		kafkaProducer:  kafkaProducer,
	}
}

func (d *OutboxDispatcher) Run(ctx context.Context, numWorkers int) <-chan error {
	errChs := make([]<-chan error, numWorkers+1)

	const overhead int = 1 // making buffer length > numWorkers to prevent workers idling while waiting db response
	genOut, genErr := d.messagesGenerator(ctx, d.cfg.NumWorkers*(overhead+1), d.cfg.RetryIn)
	errChs[0] = genErr

	for i := range numWorkers {
		errChs[i+1] = d.messagesSender(ctx, genOut)
	}

	return uniteErrors(errChs...)
}

func (d *OutboxDispatcher) messagesGenerator(ctx context.Context, bufLen int, retryIn time.Duration) (<-chan dto.OutboxMessage, <-chan error) {
	out := make(chan dto.OutboxMessage, bufLen)
	errCh := make(chan error)

	go func(ctx context.Context) {
		defer close(out)
		defer close(errCh)

		var nextRequest time.Time

		loopCtx(
			ctx,
			func(ctx context.Context) {
				nextRequest = time.Now().Add(d.cfg.DispatchInterval)
				msgs, err := d.storageService.GetOutboxMessages(ctx, bufLen-len(out), retryIn)
				if err != nil {
					errCh <- err
					return
				}
				for _, msg := range msgs {
					out <- msg
				}
				timeToWait := nextRequest.Sub(time.Now())
				if timeToWait > 0 {
					time.Sleep(timeToWait)
				}
			},
		)
	}(ctx)

	return out, errCh
}

func (d *OutboxDispatcher) messagesSender(ctx context.Context, in <-chan dto.OutboxMessage) <-chan error {
	errCh := make(chan error)

	go func(ctx context.Context) {
		defer close(errCh)

		for msg := range in {
			if ctx.Err() != nil {
				return
			}
			err := d.kafkaProducer.SendMessage(ctx, msg.Topic, msg.Payload)
			if err != nil {
				errCh <- err
				continue
			}
			err = d.storageService.DeleteOutboxMessage(ctx, msg.ID)
			if err != nil {
				errCh <- err
				continue
			}
		}
	}(ctx)

	return errCh
}

func uniteErrors(errChs ...<-chan error) <-chan error {
	var wg sync.WaitGroup
	out := make(chan error)

	for _, errCh := range errChs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for err := range errCh {
				out <- err
			}
		}()
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func loopCtx(ctx context.Context, f func(context.Context)) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			f(ctx)
		}
	}
}
