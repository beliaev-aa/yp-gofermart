package services

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	gofermartErrors "beliaev-aa/yp-gofermart/internal/gofermart/errors"
	"beliaev-aa/yp-gofermart/tests"
	"errors"
	"go.uber.org/zap"
	"testing"
)

// Тестирование метода AddOrder
func TestOrderService_AddOrder(t *testing.T) {
	testCases := []struct {
		Name          string
		Login         string
		Number        string
		MockSetup     func(m *tests.MockStorage)
		ExpectedError error
	}{
		{
			Name:   "AddOrder_Success",
			Login:  "user1",
			Number: "12345",
			MockSetup: func(m *tests.MockStorage) {
				m.GetUserByLoginFn = func(login string) (*domain.User, error) {
					return &domain.User{UserID: 1}, nil
				}
				m.GetOrderByNumberFn = func(number string) (*domain.Order, error) {
					return nil, gofermartErrors.ErrOrderNotFound
				}
				m.AddOrderFn = func(order domain.Order) error {
					return nil
				}
			},
			ExpectedError: nil,
		},
		{
			Name:   "AddOrder_ErrorInGetUserByLogin",
			Login:  "user1",
			Number: "12345",
			MockSetup: func(m *tests.MockStorage) {
				m.GetUserByLoginFn = func(login string) (*domain.User, error) {
					return nil, errors.New("failed to get user")
				}
			},
			ExpectedError: errors.New("failed to get user"),
		},
		{
			Name:   "AddOrder_ErrorInGetOrderByNumber",
			Login:  "user1",
			Number: "12345",
			MockSetup: func(m *tests.MockStorage) {
				m.GetUserByLoginFn = func(login string) (*domain.User, error) {
					return &domain.User{UserID: 1}, nil
				}
				m.GetOrderByNumberFn = func(number string) (*domain.Order, error) {
					return nil, errors.New("failed to get order by number")
				}
			},
			ExpectedError: errors.New("failed to get order by number"),
		},
		{
			Name:   "AddOrder_AlreadyUploadedByUser",
			Login:  "user1",
			Number: "12345",
			MockSetup: func(m *tests.MockStorage) {
				m.GetUserByLoginFn = func(login string) (*domain.User, error) {
					return &domain.User{UserID: 1}, nil
				}
				m.GetOrderByNumberFn = func(number string) (*domain.Order, error) {
					return &domain.Order{UserID: 1}, nil
				}
			},
			ExpectedError: gofermartErrors.ErrOrderAlreadyUploaded,
		},
		{
			Name:   "AddOrder_AlreadyUploadedByAnotherUser",
			Login:  "user2",
			Number: "12345",
			MockSetup: func(m *tests.MockStorage) {
				m.GetUserByLoginFn = func(login string) (*domain.User, error) {
					return &domain.User{UserID: 2}, nil
				}
				m.GetOrderByNumberFn = func(number string) (*domain.Order, error) {
					return &domain.Order{UserID: 1}, nil
				}
			},
			ExpectedError: gofermartErrors.ErrOrderUploadedByAnother,
		},
		{
			Name:   "AddOrder_ErrorInAddOrder",
			Login:  "user1",
			Number: "12345",
			MockSetup: func(m *tests.MockStorage) {
				m.GetUserByLoginFn = func(login string) (*domain.User, error) {
					return &domain.User{UserID: 1}, nil
				}
				m.GetOrderByNumberFn = func(number string) (*domain.Order, error) {
					return nil, gofermartErrors.ErrOrderNotFound
				}
				m.AddOrderFn = func(order domain.Order) error {
					return errors.New("failed to add order")
				}
			},
			ExpectedError: errors.New("failed to add order"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			accrualMock := &AccrualServiceMock{}
			mockStorage := &tests.MockStorage{}
			tc.MockSetup(mockStorage)

			orderService := NewOrderService(accrualMock, mockStorage, zap.NewNop())
			err := orderService.AddOrder(tc.Login, tc.Number)

			if err != nil {
				if err.Error() != tc.ExpectedError.Error() {
					t.Errorf("Expected error %v, got %v", tc.ExpectedError, err.Error())
				}
			}
		})
	}
}

// Тестирование метода GetOrders
func TestOrderService_GetOrders(t *testing.T) {
	testCases := []struct {
		Name           string
		Login          string
		MockSetup      func(m *tests.MockStorage)
		ExpectedError  string // Изменено для сравнения текста ошибки
		ExpectedOrders []domain.Order
	}{
		{
			Name:  "GetOrders_Success",
			Login: "user1",
			MockSetup: func(m *tests.MockStorage) {
				m.GetUserByLoginFn = func(login string) (*domain.User, error) {
					return &domain.User{UserID: 1}, nil
				}
				m.GetOrdersByUserIDFn = func(userID int) ([]domain.Order, error) {
					return []domain.Order{
						{OrderNumber: "12345", UserID: 1, OrderStatus: domain.OrderStatusNew},
					}, nil
				}
			},
			ExpectedError:  "",
			ExpectedOrders: []domain.Order{{OrderNumber: "12345", UserID: 1, OrderStatus: domain.OrderStatusNew}},
		},
		{
			Name:  "GetOrders_Failure_UserNotFound",
			Login: "user2",
			MockSetup: func(m *tests.MockStorage) {
				m.GetUserByLoginFn = func(login string) (*domain.User, error) {
					return nil, errors.New("user not found")
				}
			},
			ExpectedError: "user not found",
		},
		{
			Name:  "GetOrders_Failure_GetOrdersByUserIDError",
			Login: "user1",
			MockSetup: func(m *tests.MockStorage) {
				m.GetUserByLoginFn = func(login string) (*domain.User, error) {
					return &domain.User{UserID: 1}, nil
				}
				m.GetOrdersByUserIDFn = func(userID int) ([]domain.Order, error) {
					return nil, errors.New("failed to get orders by user ID")
				}
			},
			ExpectedError: "failed to get orders by user ID",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			accrualMock := &AccrualServiceMock{}
			mockStorage := &tests.MockStorage{}
			tc.MockSetup(mockStorage)

			orderService := NewOrderService(accrualMock, mockStorage, zap.NewNop())
			orders, err := orderService.GetOrders(tc.Login)

			if err != nil {
				if err.Error() != tc.ExpectedError {
					t.Errorf("Expected error %v, got %v", tc.ExpectedError, err.Error())
				}
			} else if len(orders) != len(tc.ExpectedOrders) {
				t.Errorf("Expected orders %v, got %v", tc.ExpectedOrders, orders)
			}
		})
	}
}

// Тестирование метода UpdateUserBalance
func TestOrderService_UpdateUserBalance(t *testing.T) {
	testCases := []struct {
		Name          string
		UserID        int
		Amount        float64
		MockSetup     func(m *tests.MockStorage)
		ExpectedError string // Изменено для сравнения текстов ошибок
	}{
		{
			Name:   "UpdateUserBalance_Success",
			UserID: 1,
			Amount: 100.0,
			MockSetup: func(m *tests.MockStorage) {
				m.UpdateUserBalanceFn = func(userID int, amount float64) error {
					return nil
				}
			},
			ExpectedError: "",
		},
		{
			Name:   "UpdateUserBalance_Failure",
			UserID: 1,
			Amount: 100.0,
			MockSetup: func(m *tests.MockStorage) {
				m.UpdateUserBalanceFn = func(userID int, amount float64) error {
					return errors.New("failed to update balance")
				}
			},
			ExpectedError: "failed to update balance",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			accrualMock := &AccrualServiceMock{}
			mockStorage := &tests.MockStorage{}
			tc.MockSetup(mockStorage)

			orderService := NewOrderService(accrualMock, mockStorage, zap.NewNop())
			err := orderService.UpdateUserBalance(tc.UserID, tc.Amount)

			if err != nil {
				if err.Error() != tc.ExpectedError {
					t.Errorf("Expected error %v, got %v", tc.ExpectedError, err.Error())
				}
			} else if tc.ExpectedError != "" {
				t.Errorf("Expected error %v, but got none", tc.ExpectedError)
			}
		})
	}
}

// Тестирование метода UpdateOrderStatuses
func TestOrderService_UpdateOrderStatuses(t *testing.T) {
	testCases := []struct {
		Name          string
		MockSetup     func(m *tests.MockStorage, accrualMock *AccrualServiceMock)
		CheckCalled   func(m *tests.MockStorage) error
		ExpectedError string
	}{
		{
			Name: "UpdateOrderStatuses_Success",
			MockSetup: func(m *tests.MockStorage, accrualMock *AccrualServiceMock) {
				m.GetOrdersForProcessingFn = func() ([]domain.Order, error) {
					return []domain.Order{{OrderNumber: "12345", UserID: 1}}, nil
				}
				m.LockOrderForProcessingFn = func(orderNumber string) error {
					return nil
				}
				m.UnlockOrderFn = func(orderNumber string) error {
					return nil
				}
				m.UpdateOrderFn = func(order domain.Order) error {
					return nil
				}
				m.UpdateUserBalanceFn = func(userID int, amount float64) error {
					return nil
				}
				accrualMock.GetOrderAccrualFn = func(orderNumber string) (float64, string, error) {
					return 100.0, domain.OrderStatusProcessed, nil
				}
			},
			CheckCalled: func(m *tests.MockStorage) error {
				if !m.GetOrdersForProcessingCalled {
					return errors.New("GetOrdersForProcessing was not called")
				}
				if !m.LockOrderForProcessingCalled {
					return errors.New("LockOrderForProcessing was not called")
				}
				if !m.UpdateOrderCalled {
					return errors.New("UpdateOrder was not called")
				}
				if !m.UpdateUserBalanceCalled {
					return errors.New("UpdateUserBalance was not called")
				}
				if !m.UnlockOrderCalled {
					return errors.New("UnlockOrder was not called")
				}
				return nil
			},
			ExpectedError: "",
		},
		{
			Name: "UpdateOrderStatuses_Failure_GetOrdersForProcessing",
			MockSetup: func(m *tests.MockStorage, accrualMock *AccrualServiceMock) {
				m.GetOrdersForProcessingFn = func() ([]domain.Order, error) {
					return nil, errors.New("failed to fetch orders")
				}
			},
			CheckCalled: func(m *tests.MockStorage) error {
				if !m.GetOrdersForProcessingCalled {
					return errors.New("GetOrdersForProcessing was not called")
				}
				return nil
			},
			ExpectedError: "failed to fetch orders",
		},
		{
			Name: "UpdateOrderStatuses_Failure_LockOrderForProcessing",
			MockSetup: func(m *tests.MockStorage, accrualMock *AccrualServiceMock) {
				m.GetOrdersForProcessingFn = func() ([]domain.Order, error) {
					return []domain.Order{{OrderNumber: "12345", UserID: 1}}, nil
				}
				m.LockOrderForProcessingFn = func(orderNumber string) error {
					return errors.New("failed to lock order")
				}
			},
			CheckCalled: func(m *tests.MockStorage) error {
				if !m.GetOrdersForProcessingCalled {
					return errors.New("GetOrdersForProcessing was not called")
				}
				if !m.LockOrderForProcessingCalled {
					return errors.New("LockOrderForProcessing was not called")
				}
				return nil
			},
			ExpectedError: "failed to lock order",
		},
		{
			Name: "UpdateOrderStatuses_Failure_GetOrderAccrual",
			MockSetup: func(m *tests.MockStorage, accrualMock *AccrualServiceMock) {
				m.GetOrdersForProcessingFn = func() ([]domain.Order, error) {
					return []domain.Order{{OrderNumber: "12345", UserID: 1}}, nil
				}
				m.LockOrderForProcessingFn = func(orderNumber string) error {
					return nil
				}
				accrualMock.GetOrderAccrualFn = func(orderNumber string) (float64, string, error) {
					return 0, "", errors.New("failed to fetch accrual")
				}
			},
			CheckCalled: func(m *tests.MockStorage) error {
				if !m.GetOrdersForProcessingCalled {
					return errors.New("GetOrdersForProcessing was not called")
				}
				if !m.LockOrderForProcessingCalled {
					return errors.New("LockOrderForProcessing was not called")
				}
				return nil
			},
			ExpectedError: "failed to fetch accrual",
		},
		{
			Name: "UpdateOrderStatuses_Failure_UpdateOrder",
			MockSetup: func(m *tests.MockStorage, accrualMock *AccrualServiceMock) {
				m.GetOrdersForProcessingFn = func() ([]domain.Order, error) {
					return []domain.Order{{OrderNumber: "12345", UserID: 1}}, nil
				}
				m.LockOrderForProcessingFn = func(orderNumber string) error {
					return nil
				}
				accrualMock.GetOrderAccrualFn = func(orderNumber string) (float64, string, error) {
					return 100.0, domain.OrderStatusProcessed, nil
				}
				m.UpdateOrderFn = func(order domain.Order) error {
					return errors.New("failed to update order")
				}
			},
			CheckCalled: func(m *tests.MockStorage) error {
				if !m.GetOrdersForProcessingCalled {
					return errors.New("GetOrdersForProcessing was not called")
				}
				if !m.LockOrderForProcessingCalled {
					return errors.New("LockOrderForProcessing was not called")
				}
				if !m.UpdateOrderCalled {
					return errors.New("UpdateOrder was not called")
				}
				return nil
			},
			ExpectedError: "failed to update order",
		},
		{
			Name: "UpdateOrderStatuses_Failure_UpdateUserBalance",
			MockSetup: func(m *tests.MockStorage, accrualMock *AccrualServiceMock) {
				m.GetOrdersForProcessingFn = func() ([]domain.Order, error) {
					return []domain.Order{{OrderNumber: "12345", UserID: 1}}, nil
				}
				m.LockOrderForProcessingFn = func(orderNumber string) error {
					return nil
				}
				accrualMock.GetOrderAccrualFn = func(orderNumber string) (float64, string, error) {
					return 100.0, domain.OrderStatusProcessed, nil
				}
				m.UpdateOrderFn = func(order domain.Order) error {
					return nil
				}
				m.UpdateUserBalanceFn = func(userID int, amount float64) error {
					return errors.New("failed to update balance")
				}
			},
			CheckCalled: func(m *tests.MockStorage) error {
				if !m.GetOrdersForProcessingCalled {
					return errors.New("GetOrdersForProcessing was not called")
				}
				if !m.LockOrderForProcessingCalled {
					return errors.New("LockOrderForProcessing was not called")
				}
				if !m.UpdateOrderCalled {
					return errors.New("UpdateOrder was not called")
				}
				if !m.UpdateUserBalanceCalled {
					return errors.New("UpdateUserBalance was not called")
				}
				return nil
			},
			ExpectedError: "failed to update balance",
		},
		{
			Name: "UpdateOrderStatuses_Failure_UnlockOrder",
			MockSetup: func(m *tests.MockStorage, accrualMock *AccrualServiceMock) {
				m.GetOrdersForProcessingFn = func() ([]domain.Order, error) {
					return []domain.Order{{OrderNumber: "12345", UserID: 1}}, nil
				}
				m.LockOrderForProcessingFn = func(orderNumber string) error {
					return nil
				}
				accrualMock.GetOrderAccrualFn = func(orderNumber string) (float64, string, error) {
					return 100.0, domain.OrderStatusProcessed, nil
				}
				m.UpdateOrderFn = func(order domain.Order) error {
					return nil
				}
				m.UpdateUserBalanceFn = func(userID int, amount float64) error {
					return nil
				}
				m.UnlockOrderFn = func(orderNumber string) error {
					return errors.New("failed to unlock order")
				}
			},
			CheckCalled: func(m *tests.MockStorage) error {
				if !m.GetOrdersForProcessingCalled {
					return errors.New("GetOrdersForProcessing was not called")
				}
				if !m.LockOrderForProcessingCalled {
					return errors.New("LockOrderForProcessing was not called")
				}
				if !m.UpdateOrderCalled {
					return errors.New("UpdateOrder was not called")
				}
				if !m.UpdateUserBalanceCalled {
					return errors.New("UpdateUserBalance was not called")
				}
				if !m.UnlockOrderCalled {
					return errors.New("UnlockOrder was not called")
				}
				return nil
			},
			ExpectedError: "failed to unlock order",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mockStorage := &tests.MockStorage{}
			accrualMock := &AccrualServiceMock{}
			tc.MockSetup(mockStorage, accrualMock)

			// Передаем mock в OrderService
			orderService := NewOrderService(accrualMock, mockStorage, zap.NewNop())
			orderService.UpdateOrderStatuses()

			// Проверяем, были ли вызваны ожидаемые методы
			if err := tc.CheckCalled(mockStorage); err != nil {
				t.Errorf("Method call mismatch: %v", err)
			}
		})
	}
}

// AccrualServiceMock - это мок для внешнего сервиса начислений
type AccrualServiceMock struct {
	GetOrderAccrualFn func(orderNumber string) (float64, string, error)
}

func (m *AccrualServiceMock) GetOrderAccrual(orderNumber string) (float64, string, error) {
	return m.GetOrderAccrualFn(orderNumber)
}
