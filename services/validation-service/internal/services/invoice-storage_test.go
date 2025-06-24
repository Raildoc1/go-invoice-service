package services

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go-invoice-service/common/pkg/logging"
	"go-invoice-service/common/protocol/proto/types"
	"go-invoice-service/common/protocol/proto/validation"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/timestamppb"
	"testing"
	"validation-service/internal/dto"
	mock_services "validation-service/internal/services/mocks"
)

func TestInvoiceStorage_GetInvoice(t *testing.T) {
	tests := []struct {
		name           string
		storageErr     error
		invoiceFactory func(uuid.UUID) *types.Invoice
		receivedStatus types.InvoiceStatus
		resultCheck    func(*testing.T, *dto.Invoice, dto.InvoiceStatus, error)
	}{
		{
			name:           "success_pending",
			invoiceFactory: createInvoice,
			receivedStatus: types.InvoiceStatus_Pending,
			resultCheck: func(t *testing.T, invoice *dto.Invoice, status dto.InvoiceStatus, err error) {
				require.NoError(t, err)
				assert.Equal(t, dto.PendingInvoiceStatus, status)
			},
		},
		{
			name:           "success_approved",
			invoiceFactory: createInvoice,
			receivedStatus: types.InvoiceStatus_Approved,
			resultCheck: func(t *testing.T, invoice *dto.Invoice, status dto.InvoiceStatus, err error) {
				require.NoError(t, err)
				assert.Equal(t, dto.ApprovedInvoiceStatus, status)
			},
		},
		{
			name:           "success_rejected",
			invoiceFactory: createInvoice,
			receivedStatus: types.InvoiceStatus_Rejected,
			resultCheck: func(t *testing.T, invoice *dto.Invoice, status dto.InvoiceStatus, err error) {
				require.NoError(t, err)
				assert.Equal(t, dto.RejectedInvoiceStatus, status)
			},
		},
		{
			name:           "invalid_status",
			invoiceFactory: createInvoice,
			receivedStatus: types.InvoiceStatus(-1),
			resultCheck: func(t *testing.T, invoice *dto.Invoice, status dto.InvoiceStatus, err error) {
				require.Error(t, err)
				assert.Nil(t, invoice)
				assert.Equal(t, dto.NilInvoiceStatus, status)
			},
		},
		{
			name: "invalid_invoice_uuid_received_from_storage",
			invoiceFactory: func(u uuid.UUID) *types.Invoice {
				invoice := createInvoice(u)
				invoice.Id = &types.UUID{
					Value: invoice.Id.Value[:8],
				}
				return invoice
			},
			receivedStatus: types.InvoiceStatus_Rejected,
			resultCheck: func(t *testing.T, invoice *dto.Invoice, status dto.InvoiceStatus, err error) {
				require.Error(t, err)
				assert.Nil(t, invoice)
				assert.Equal(t, dto.NilInvoiceStatus, status)
			},
		},
		{
			name: "invalid_customer_uuid_received_from_storage",
			invoiceFactory: func(u uuid.UUID) *types.Invoice {
				invoice := createInvoice(u)
				invoice.CustomerId = &types.UUID{
					Value: invoice.CustomerId.Value[:8],
				}
				return invoice
			},
			receivedStatus: types.InvoiceStatus_Rejected,
			resultCheck: func(t *testing.T, invoice *dto.Invoice, status dto.InvoiceStatus, err error) {
				require.Error(t, err)
				assert.Nil(t, invoice)
				assert.Equal(t, dto.NilInvoiceStatus, status)
			},
		},
		{
			name:       "storage_err",
			storageErr: errors.New("some error"),
			resultCheck: func(t *testing.T, invoice *dto.Invoice, status dto.InvoiceStatus, err error) {
				require.Error(t, err)
				assert.Nil(t, invoice)
				assert.Equal(t, dto.NilInvoiceStatus, status)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctl := gomock.NewController(t)
			defer ctl.Finish()

			invoiceStorageClient := mock_services.NewMockInvoiceStorageClient(ctl)
			invoicesMetrics := mock_services.NewMockInvoicesMetrics(ctl)
			logger := logging.NewNopLogger()

			invoiceStorage := NewInvoiceStorage(invoiceStorageClient, invoicesMetrics, logger)

			invoiceID := uuid.New()
			request := &validation.GetInvoiceRequest{
				Id: &types.UUID{
					Value: invoiceID[:],
				},
			}

			if test.storageErr == nil {
				invoice := test.invoiceFactory(invoiceID)
				invoiceStorageClient.EXPECT().
					Get(gomock.Any(), request).
					Return(&validation.GetInvoiceResponse{
						Invoice: invoice,
						Status:  &test.receivedStatus,
					}, nil).
					Times(1)
			} else {
				invoiceStorageClient.EXPECT().
					Get(gomock.Any(), request).
					Return(nil, test.storageErr).
					Times(1)
			}

			resInvoice, resStatus, resErr := invoiceStorage.GetInvoice(context.Background(), invoiceID)
			test.resultCheck(t, resInvoice, resStatus, resErr)
		})
	}
}

func createInvoice(invoiceID uuid.UUID) *types.Invoice {
	customerID := uuid.New()
	amount := int64(10000)
	currency := "USD"
	notes := "some_notes"

	description := "some description"
	quantity := int32(10000)
	unitPrice := int64(10000)
	total := int64(quantity) * unitPrice

	return &types.Invoice{
		Id: &types.UUID{
			Value: invoiceID[:],
		},
		CustomerId: &types.UUID{
			Value: customerID[:],
		},
		Amount:    &amount,
		Currency:  &currency,
		DueDate:   &timestamppb.Timestamp{Seconds: 15},
		CreatedAt: &timestamppb.Timestamp{Seconds: 15},
		UpdatedAt: &timestamppb.Timestamp{Seconds: 15},
		Items: []*types.Item{
			{
				Description: &description,
				Quantity:    &quantity,
				UnitPrice:   &unitPrice,
				Total:       &total,
			},
		},
		Notes: &notes,
	}
}

func TestInvoiceStorage_SetApproved(t *testing.T) {
	tests := []struct {
		name                   string
		errFromStorage         error
		metricsCollectionTimes int
		resultCheck            func(*testing.T, error)
	}{
		{
			name:                   "success",
			errFromStorage:         nil,
			metricsCollectionTimes: 1,
			resultCheck: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name:                   "fail",
			errFromStorage:         errors.New("test storage error"),
			metricsCollectionTimes: 0,
			resultCheck: func(t *testing.T, err error) {
				require.Error(t, err)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			invoiceStorageClient := mock_services.NewMockInvoiceStorageClient(ctrl)
			invoicesMetrics := mock_services.NewMockInvoicesMetrics(ctrl)
			logger := logging.NewNopLogger()

			invoiceStorage := NewInvoiceStorage(invoiceStorageClient, invoicesMetrics, logger)

			invoiceID := uuid.New()
			request := &validation.SetApprovedRequest{
				Id: &types.UUID{
					Value: invoiceID[:],
				},
			}
			invoiceStorageClient.EXPECT().
				SetApproved(gomock.Any(), request).
				Return(nil, test.errFromStorage).
				Times(1)

			invoicesMetrics.EXPECT().
				IncTotalHandledInvoices(gomock.Any(), "approved").
				Times(test.metricsCollectionTimes)

			err := invoiceStorage.SetApproved(context.Background(), invoiceID)
			test.resultCheck(t, err)
		})
	}
}

func TestInvoiceStorage_SetRejected(t *testing.T) {
	tests := []struct {
		name                   string
		errFromStorage         error
		metricsCollectionTimes int
		resultCheck            func(*testing.T, error)
	}{
		{
			name:                   "success",
			errFromStorage:         nil,
			metricsCollectionTimes: 1,
			resultCheck: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name:                   "fail",
			errFromStorage:         errors.New("test storage error"),
			metricsCollectionTimes: 0,
			resultCheck: func(t *testing.T, err error) {
				require.Error(t, err)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			invoiceStorageClient := mock_services.NewMockInvoiceStorageClient(ctrl)
			invoicesMetrics := mock_services.NewMockInvoicesMetrics(ctrl)
			logger := logging.NewNopLogger()

			invoiceStorage := NewInvoiceStorage(invoiceStorageClient, invoicesMetrics, logger)

			invoiceID := uuid.New()
			request := &validation.SetRejectedRequest{
				Id: &types.UUID{
					Value: invoiceID[:],
				},
			}
			invoiceStorageClient.EXPECT().
				SetRejected(gomock.Any(), request).
				Return(nil, test.errFromStorage).
				Times(1)

			invoicesMetrics.EXPECT().
				IncTotalHandledInvoices(gomock.Any(), "rejected").
				Times(test.metricsCollectionTimes)

			err := invoiceStorage.SetRejected(context.Background(), invoiceID)
			test.resultCheck(t, err)
		})
	}
}
