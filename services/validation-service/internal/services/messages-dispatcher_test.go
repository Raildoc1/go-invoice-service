package services

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go-invoice-service/common/pkg/logging"
	protocol "go-invoice-service/common/protocol/kafka"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
	"validation-service/internal/dto"
	mock_services "validation-service/internal/services/mocks"
)

func TestMessagesDispatcher_HandleMessage(t *testing.T) {
	invoiceID := uuid.New()

	tests := []struct {
		name                  string
		messageBody           []byte
		invalidBody           bool
		invoiceProvider       func() *dto.Invoice
		invoiceStatus         dto.InvoiceStatus
		storageError          error
		validateResult        bool
		storageStatusSetError error
		resultCheck           func(*testing.T, error)
	}{
		{
			name:        "success_pending",
			messageBody: messageFromId(invoiceID),
			invoiceProvider: func() *dto.Invoice {
				return createInvoiceDTO(invoiceID)
			},
			invoiceStatus:  dto.PendingInvoiceStatus,
			storageError:   nil,
			validateResult: true,
			resultCheck: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name:        "success_approved_skipped",
			messageBody: messageFromId(invoiceID),
			invoiceProvider: func() *dto.Invoice {
				return createInvoiceDTO(invoiceID)
			},
			invoiceStatus:  dto.ApprovedInvoiceStatus,
			storageError:   nil,
			validateResult: true,
			resultCheck: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name:        "success_rejected_skipped",
			messageBody: messageFromId(invoiceID),
			invoiceProvider: func() *dto.Invoice {
				return createInvoiceDTO(invoiceID)
			},
			invoiceStatus:  dto.RejectedInvoiceStatus,
			storageError:   nil,
			validateResult: true,
			resultCheck: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name:        "success_rejected",
			messageBody: messageFromId(invoiceID),
			invoiceProvider: func() *dto.Invoice {
				return createInvoiceDTO(invoiceID)
			},
			invoiceStatus:  dto.PendingInvoiceStatus,
			storageError:   nil,
			validateResult: false,
			resultCheck: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name:        "storage_error",
			messageBody: messageFromId(invoiceID),
			invoiceProvider: func() *dto.Invoice {
				return nil
			},
			invoiceStatus:  dto.NilInvoiceStatus,
			storageError:   errors.New("test storage error"),
			validateResult: true,
			resultCheck: func(t *testing.T, err error) {
				require.Error(t, err)
			},
		},
		{
			name:        "set_approved_error",
			messageBody: messageFromId(invoiceID),
			invoiceProvider: func() *dto.Invoice {
				return createInvoiceDTO(invoiceID)
			},
			invoiceStatus:         dto.PendingInvoiceStatus,
			validateResult:        true,
			storageStatusSetError: errors.New("test storage error"),
			resultCheck: func(t *testing.T, err error) {
				require.Error(t, err)
			},
		},
		{
			name:        "set_rejected_error",
			messageBody: messageFromId(invoiceID),
			invoiceProvider: func() *dto.Invoice {
				return createInvoiceDTO(invoiceID)
			},
			invoiceStatus:         dto.PendingInvoiceStatus,
			validateResult:        false,
			storageStatusSetError: errors.New("test storage error"),
			resultCheck: func(t *testing.T, err error) {
				require.Error(t, err)
			},
		},
		{
			name:        "invalid_message_error",
			messageBody: []byte(`}invalid JSON {`),
			invalidBody: true,
			invoiceProvider: func() *dto.Invoice {
				return nil
			},
			invoiceStatus:         dto.PendingInvoiceStatus,
			validateResult:        false,
			storageStatusSetError: errors.New("test storage error"),
			resultCheck: func(t *testing.T, err error) {
				require.Error(t, err)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctl := gomock.NewController(t)
			defer ctl.Finish()

			cfg := MessagesDispatcherConfig{
				PollTimeoutMs: 100,
			}

			invoiceStorage := mock_services.NewMockInvoiceStorage(ctl)
			messagesConsumer := mock_services.NewMockMessageConsumer(ctl)
			invoiceValidator := mock_services.NewMockInvoiceValidator(ctl)
			logger := logging.NewNopLogger()

			if !test.invalidBody {
				var invoiceStorageGetInvoiceCall *gomock.Call

				if test.storageError != nil {
					invoiceStorageGetInvoiceCall = invoiceStorage.EXPECT().
						GetInvoice(gomock.Any(), gomock.Any()).
						Return(nil, dto.NilInvoiceStatus, test.storageError).
						Times(1)
				} else {
					invoiceStorageGetInvoiceCall = invoiceStorage.EXPECT().
						GetInvoice(gomock.Any(), gomock.Any()).
						Return(test.invoiceProvider(), test.invoiceStatus, nil).
						Times(1)

					if test.invoiceStatus == dto.PendingInvoiceStatus {
						invoiceValidator.EXPECT().
							Validate(gomock.Any()).
							Return(test.validateResult).
							After(invoiceStorageGetInvoiceCall).
							Times(1)

						if test.validateResult {
							invoiceStorage.EXPECT().
								SetApproved(gomock.Any(), gomock.Any()).
								Return(test.storageStatusSetError).
								Times(1)
						} else {
							invoiceStorage.EXPECT().
								SetRejected(gomock.Any(), gomock.Any()).
								Return(test.storageStatusSetError).
								Times(1)
						}
					}
				}
			}

			messagesDispatcher := NewMessagesDispatcher(
				cfg,
				invoiceStorage,
				messagesConsumer,
				invoiceValidator,
				logger,
			)

			err := messagesDispatcher.HandleMessage(context.Background(), test.messageBody)
			test.resultCheck(t, err)
		})
	}
}

func messageFromId(id uuid.UUID) []byte {
	val := protocol.NewInvoice{
		ID: id,
	}
	res, err := json.Marshal(val)
	if err != nil {
		panic(err)
	}
	return res
}

func createInvoiceDTO(id uuid.UUID) *dto.Invoice {
	return &dto.Invoice{
		ID:         id,
		CustomerID: uuid.New(),
		Amount:     1000,
		Currency:   "USD",
		DueDate:    time.Now(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Items: []dto.Item{
			{
				Description: "Item 1",
				Quantity:    10,
				UnitPrice:   100,
				Total:       1000,
			},
		},
		Notes: "some notes",
	}
}

func TestMessagesDispatcher_Tick(t *testing.T) {
	tests := []struct {
		name               string
		peekNext           func() ([]byte, error)
		handleMessageError error
		commitError        error
		resultCheck        func(*testing.T, error)
	}{
		{
			name: "success",
			peekNext: func() ([]byte, error) {
				return []byte{1, 2, 3}, nil
			},
			resultCheck: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "peek_fail",
			peekNext: func() ([]byte, error) {
				return nil, errors.New("peek fail")
			},
			resultCheck: func(t *testing.T, err error) {
				require.Error(t, err)
			},
		},
		{
			name: "commit_fail",
			peekNext: func() ([]byte, error) {
				return []byte{1, 2, 3}, nil
			},
			commitError: errors.New("commit fail"),
			resultCheck: func(t *testing.T, err error) {
				require.Error(t, err)
			},
		},
		{
			name: "handle_message_fail",
			peekNext: func() ([]byte, error) {
				return []byte{1, 2, 3}, nil
			},
			handleMessageError: errors.New("handle message fail"),
			resultCheck: func(t *testing.T, err error) {
				require.Error(t, err)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctl := gomock.NewController(t)
			defer ctl.Finish()

			cfg := MessagesDispatcherConfig{
				PollTimeoutMs: 100,
			}

			invoiceStorage := mock_services.NewMockInvoiceStorage(ctl)
			messagesConsumer := mock_services.NewMockMessageConsumer(ctl)
			invoiceValidator := mock_services.NewMockInvoiceValidator(ctl)
			logger := logging.NewNopLogger()

			peekCall := messagesConsumer.EXPECT().
				PeekNext(cfg.PollTimeoutMs).
				Return(test.peekNext()).
				Times(1)

			messagesConsumer.EXPECT().
				Commit(gomock.Any()).
				Return(test.commitError).
				MaxTimes(1).
				After(peekCall)

			messagesDispatcher := NewMessagesDispatcher(
				cfg,
				invoiceStorage,
				messagesConsumer,
				invoiceValidator,
				logger,
			)

			err := messagesDispatcher.tick(
				context.Background(),
				func(ctx context.Context, bytes []byte) error {
					return test.handleMessageError
				},
			)
			test.resultCheck(t, err)

		})
	}
}
