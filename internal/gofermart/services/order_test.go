package services

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	gofermartErrors "beliaev-aa/yp-gofermart/internal/gofermart/errors"
	"beliaev-aa/yp-gofermart/tests/mocks"
	"bytes"
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm"
	"testing"
)

func TestOrderService_AddOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrderRepo := mocks.NewMockOrderRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)

	logger := zap.NewNop()
	orderService := NewOrderService(nil, mockOrderRepo, mockUserRepo, logger)

	testCases := []struct {
		name          string
		login         string
		orderNumber   string
		setupMocks    func()
		expectedError error
	}{
		{
			name:        "User_Not_Found",
			login:       "user1",
			orderNumber: "123456789",
			setupMocks: func() {
				mockUserRepo.EXPECT().GetUserByLogin(gomock.Any(), "user1").Return(nil, gofermartErrors.ErrUserNotFound)
			},
			expectedError: gofermartErrors.ErrUserNotFound,
		},
		{
			name:        "GetOrderByNumber_Unexpected_Error",
			login:       "user1",
			orderNumber: "123456789",
			setupMocks: func() {
				mockUserRepo.EXPECT().GetUserByLogin(gomock.Any(), "user1").Return(&domain.User{UserID: 1}, nil)
				mockOrderRepo.EXPECT().GetOrderByNumber(gomock.Any(), "123456789").Return(nil, errors.New("unexpected error"))
			},
			expectedError: errors.New("unexpected error"),
		},
		{
			name:        "Order_Already_Uploaded_Same_User",
			login:       "user1",
			orderNumber: "123456789",
			setupMocks: func() {
				mockUserRepo.EXPECT().GetUserByLogin(gomock.Any(), "user1").Return(&domain.User{UserID: 1}, nil)
				mockOrderRepo.EXPECT().GetOrderByNumber(gomock.Any(), "123456789").Return(&domain.Order{UserID: 1}, nil)
			},
			expectedError: gofermartErrors.ErrOrderAlreadyUploaded,
		},
		{
			name:        "Order_Already_Uploaded_Different_User",
			login:       "user1",
			orderNumber: "123456789",
			setupMocks: func() {
				mockUserRepo.EXPECT().GetUserByLogin(gomock.Any(), "user1").Return(&domain.User{UserID: 1}, nil)
				mockOrderRepo.EXPECT().GetOrderByNumber(gomock.Any(), "123456789").Return(&domain.Order{UserID: 2}, nil)
			},
			expectedError: gofermartErrors.ErrOrderUploadedByAnother,
		},
		{
			name:        "Order_Not_Found_Add_Success",
			login:       "user1",
			orderNumber: "123456789",
			setupMocks: func() {
				mockUserRepo.EXPECT().GetUserByLogin(gomock.Any(), "user1").Return(&domain.User{UserID: 1}, nil)
				mockOrderRepo.EXPECT().GetOrderByNumber(gomock.Any(), "123456789").Return(nil, gofermartErrors.ErrOrderNotFound)
				mockOrderRepo.EXPECT().AddOrder(gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:        "Add_Order_Failure",
			login:       "user1",
			orderNumber: "123456789",
			setupMocks: func() {
				mockUserRepo.EXPECT().GetUserByLogin(gomock.Any(), "user1").Return(&domain.User{UserID: 1}, nil)
				mockOrderRepo.EXPECT().GetOrderByNumber(gomock.Any(), "123456789").Return(nil, gofermartErrors.ErrOrderNotFound)
				mockOrderRepo.EXPECT().AddOrder(gomock.Any(), gomock.Any()).Return(errors.New("db error"))
			},
			expectedError: errors.New("db error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()

			err := orderService.AddOrder(tc.login, tc.orderNumber)

			if err != nil && tc.expectedError == nil {
				t.Errorf("Expected no error, got %v", err)
			} else if err == nil && tc.expectedError != nil {
				t.Errorf("Expected error, got none")
			} else if err != nil && err.Error() != tc.expectedError.Error() {
				t.Errorf("Expected error %v, got %v", tc.expectedError, err)
			}
		})
	}
}

func TestOrderService_GetOrders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrderRepo := mocks.NewMockOrderRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	logger := zap.NewNop()
	orderService := NewOrderService(nil, mockOrderRepo, mockUserRepo, logger)

	testCases := []struct {
		name           string
		login          string
		setupMocks     func()
		expectedError  error
		expectedOrders []domain.Order
	}{
		{
			name:  "User_Not_Found",
			login: "user1",
			setupMocks: func() {
				mockUserRepo.EXPECT().GetUserByLogin(gomock.Any(), "user1").Return(nil, gofermartErrors.ErrUserNotFound)
			},
			expectedError:  gofermartErrors.ErrUserNotFound,
			expectedOrders: nil,
		},
		{
			name:  "GetOrdersByUserID_Error",
			login: "user1",
			setupMocks: func() {
				mockUserRepo.EXPECT().GetUserByLogin(gomock.Any(), "user1").Return(&domain.User{UserID: 1}, nil)
				mockOrderRepo.EXPECT().GetOrdersByUserID(gomock.Any(), 1).Return(nil, errors.New("db error"))
			},
			expectedError:  errors.New("db error"),
			expectedOrders: nil,
		},
		{
			name:  "GetOrders_Success",
			login: "user1",
			setupMocks: func() {
				mockUserRepo.EXPECT().GetUserByLogin(gomock.Any(), "user1").Return(&domain.User{UserID: 1}, nil)
				mockOrderRepo.EXPECT().GetOrdersByUserID(gomock.Any(), 1).Return([]domain.Order{
					{OrderNumber: "123456789", UserID: 1, OrderStatus: domain.OrderStatusNew},
					{OrderNumber: "987654321", UserID: 1, OrderStatus: domain.OrderStatusProcessed},
				}, nil)
			},
			expectedError: nil,
			expectedOrders: []domain.Order{
				{OrderNumber: "123456789", UserID: 1, OrderStatus: domain.OrderStatusNew},
				{OrderNumber: "987654321", UserID: 1, OrderStatus: domain.OrderStatusProcessed},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()

			orders, err := orderService.GetOrders(tc.login)

			if err != nil && tc.expectedError == nil {
				t.Errorf("Expected no error, got %v", err)
			} else if err == nil && tc.expectedError != nil {
				t.Errorf("Expected error, got none")
			} else if err != nil && err.Error() != tc.expectedError.Error() {
				t.Errorf("Expected error %v, got %v", tc.expectedError, err)
			}

			if len(orders) != len(tc.expectedOrders) {
				diff := cmp.Diff(tc.expectedOrders, orders)
				t.Errorf("expected orders mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestOrderService_UpdateOrderStatuses(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrderRepo := mocks.NewMockOrderRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockAccrualClient := mocks.NewMockAccrualService(ctrl)

	var logBuffer bytes.Buffer
	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(&logBuffer),
		zapcore.DebugLevel,
	))

	orderService := NewOrderService(mockAccrualClient, mockOrderRepo, mockUserRepo, logger)

	testCases := []struct {
		name        string
		setupMocks  func()
		expectedLog string
	}{
		{
			name: "Successful_Order_Processing",
			setupMocks: func() {
				mockOrderRepo.EXPECT().GetOrdersForProcessing(gomock.Any()).Return([]domain.Order{
					{OrderNumber: "order123", UserID: 1, OrderStatus: domain.OrderStatusNew},
				}, nil)
				mockUserRepo.EXPECT().BeginTransaction().Return(&gorm.DB{}, nil)
				mockAccrualClient.EXPECT().GetOrderAccrual(gomock.Any(), "order123").Return(100.0, domain.OrderStatusProcessed, nil)
				mockOrderRepo.EXPECT().UpdateOrder(gomock.Any(), gomock.Any()).Return(nil)
				mockUserRepo.EXPECT().UpdateUserBalance(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				mockOrderRepo.EXPECT().UnlockOrder(gomock.Any(), "order123").Return(nil)
				mockUserRepo.EXPECT().Commit(gomock.Any()).Return(nil)
			},
			expectedLog: `"msg":"Order processed successfully"`,
		},
		{
			name: "Failed_To_Fetch_Orders_For_Processing",
			setupMocks: func() {
				mockOrderRepo.EXPECT().GetOrdersForProcessing(gomock.Any()).Return(nil, errors.New("failed to fetch orders"))
			},
			expectedLog: `"msg":"Failed to fetch orders for status update"`,
		},
		{
			name: "Failed_To_Start_Transaction",
			setupMocks: func() {
				mockOrderRepo.EXPECT().GetOrdersForProcessing(gomock.Any()).Return([]domain.Order{
					{OrderNumber: "order123", UserID: 1, OrderStatus: domain.OrderStatusNew},
				}, nil)
				mockUserRepo.EXPECT().BeginTransaction().Return(nil, errors.New("failed to start transaction"))
			},
			expectedLog: `"msg":"Failed to start transaction"`,
		},
		{
			name: "Failed_To_Commit_Transaction",
			setupMocks: func() {
				mockOrderRepo.EXPECT().GetOrdersForProcessing(gomock.Any()).Return([]domain.Order{
					{OrderNumber: "order123", UserID: 1, OrderStatus: domain.OrderStatusNew},
				}, nil)
				mockUserRepo.EXPECT().BeginTransaction().Return(&gorm.DB{}, nil)
				mockAccrualClient.EXPECT().GetOrderAccrual(gomock.Any(), "order123").Return(100.0, domain.OrderStatusProcessed, nil)
				mockOrderRepo.EXPECT().UpdateOrder(gomock.Any(), gomock.Any()).Return(nil)
				mockUserRepo.EXPECT().UpdateUserBalance(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				mockOrderRepo.EXPECT().UnlockOrder(gomock.Any(), "order123").Return(nil)
				mockUserRepo.EXPECT().Commit(gomock.Any()).Return(errors.New("failed to commit transaction"))
			},
			expectedLog: `"msg":"Failed to commit transaction"`,
		},
		{
			name: "Failed_To_Rollback_Transaction",
			setupMocks: func() {
				mockOrderRepo.EXPECT().GetOrdersForProcessing(gomock.Any()).Return([]domain.Order{
					{OrderNumber: "order123", UserID: 1, OrderStatus: domain.OrderStatusNew},
				}, nil)
				mockUserRepo.EXPECT().BeginTransaction().Return(&gorm.DB{}, nil)
				mockAccrualClient.EXPECT().GetOrderAccrual(gomock.Any(), "order123").Return(0.0, domain.OrderStatusNew, errors.New("accrual error"))
				mockUserRepo.EXPECT().Rollback(gomock.Any()).Return(errors.New("failed to rollback transaction"))
			},
			expectedLog: `"msg":"Failed to rollback transaction"`,
		},
		{
			name: "Failed_To_Update_Order",
			setupMocks: func() {
				mockOrderRepo.EXPECT().GetOrdersForProcessing(gomock.Any()).Return([]domain.Order{
					{OrderNumber: "order123", UserID: 1, OrderStatus: domain.OrderStatusNew},
				}, nil)
				mockUserRepo.EXPECT().BeginTransaction().Return(&gorm.DB{}, nil)
				mockAccrualClient.EXPECT().GetOrderAccrual(gomock.Any(), "order123").Return(100.0, domain.OrderStatusNew, nil)
				mockOrderRepo.EXPECT().UpdateOrder(gomock.Any(), gomock.Any()).Return(errors.New("failed to update order"))
				mockUserRepo.EXPECT().Rollback(gomock.Any()).Return(nil)
			},
			expectedLog: `"msg":"Failed to update order"`,
		},
		{
			name: "Failed_To_Update_User_Balance",
			setupMocks: func() {
				mockOrderRepo.EXPECT().GetOrdersForProcessing(gomock.Any()).Return([]domain.Order{
					{OrderNumber: "order123", UserID: 1, OrderStatus: domain.OrderStatusNew},
				}, nil)
				mockUserRepo.EXPECT().BeginTransaction().Return(&gorm.DB{}, nil)
				mockAccrualClient.EXPECT().GetOrderAccrual(gomock.Any(), "order123").Return(100.0, domain.OrderStatusProcessed, nil)
				mockOrderRepo.EXPECT().UpdateOrder(gomock.Any(), gomock.Any()).Return(nil)
				mockUserRepo.EXPECT().UpdateUserBalance(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("failed to update user balance"))
				mockUserRepo.EXPECT().Rollback(gomock.Any()).Return(nil)
			},
			expectedLog: `"msg":"Failed to update user balance"`,
		},
		{
			name: "Failed_To_Unlock_Order",
			setupMocks: func() {
				mockOrderRepo.EXPECT().GetOrdersForProcessing(gomock.Any()).Return([]domain.Order{
					{OrderNumber: "order123", UserID: 1, OrderStatus: domain.OrderStatusProcessed},
				}, nil)
				mockUserRepo.EXPECT().BeginTransaction().Return(&gorm.DB{}, nil)
				mockAccrualClient.EXPECT().GetOrderAccrual(gomock.Any(), "order123").Return(100.0, domain.OrderStatusProcessed, nil)
				mockOrderRepo.EXPECT().UpdateOrder(gomock.Any(), gomock.Any()).Return(nil)
				mockUserRepo.EXPECT().UpdateUserBalance(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				mockOrderRepo.EXPECT().UnlockOrder(gomock.Any(), "order123").Return(errors.New("failed to unlock order"))
				mockUserRepo.EXPECT().Rollback(gomock.Any()).Return(nil)
			},
			expectedLog: `"msg":"Failed to unlock order"`,
		},
		{
			name: "Order_Processing_Failure",
			setupMocks: func() {
				mockOrderRepo.EXPECT().GetOrdersForProcessing(gomock.Any()).Return([]domain.Order{
					{OrderNumber: "order123", UserID: 1, OrderStatus: domain.OrderStatusNew},
				}, nil)
				mockUserRepo.EXPECT().BeginTransaction().Return(&gorm.DB{}, nil)
				mockAccrualClient.EXPECT().GetOrderAccrual(gomock.Any(), "order123").Return(0.0, domain.OrderStatusNew, errors.New("accrual error"))
				mockUserRepo.EXPECT().Rollback(gomock.Any()).Return(nil)
			},
			expectedLog: `"msg":"Failed to fetch order accrual"`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logBuffer.Reset()

			tc.setupMocks()
			orderService.UpdateOrderStatuses(context.Background())

			logOutput := logBuffer.String()
			if !bytes.Contains([]byte(logOutput), []byte(tc.expectedLog)) {
				t.Errorf("Expected log output to contain %q, but got %q", tc.expectedLog, logOutput)
			}
		})
	}
}
