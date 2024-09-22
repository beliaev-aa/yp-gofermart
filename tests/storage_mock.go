package tests

import "beliaev-aa/yp-gofermart/internal/gofermart/domain"

type MockStorage struct {
	AddOrderFn               func(order domain.Order) error
	AddWithdrawalFn          func(withdrawal domain.Withdrawal) error
	GetOrderByNumberFn       func(number string) (*domain.Order, error)
	GetOrdersForProcessingFn func() ([]domain.Order, error)
	GetOrdersByUserIDFn      func(userID int) ([]domain.Order, error)
	GetUserByLoginFn         func(login string) (*domain.User, error)
	GetWithdrawalsByUserIDFn func(userID int) ([]domain.Withdrawal, error)
	SaveUserFn               func(user domain.User) error
	UpdateOrderFn            func(order domain.Order) error
	UpdateUserBalanceFn      func(userID int, amount float64) error
	LockOrderForProcessingFn func(orderNumber string) error
	UnlockOrderFn            func(orderNumber string) error
	GetUserBalanceFn         func(login string) (float64, float64, error)
}

func (m *MockStorage) AddOrder(order domain.Order) error {
	return m.AddOrderFn(order)
}

func (m *MockStorage) AddWithdrawal(withdrawal domain.Withdrawal) error {
	return m.AddWithdrawalFn(withdrawal)
}

func (m *MockStorage) GetOrderByNumber(number string) (*domain.Order, error) {
	return m.GetOrderByNumberFn(number)
}

func (m *MockStorage) GetOrdersForProcessing() ([]domain.Order, error) {
	return m.GetOrdersForProcessingFn()
}

func (m *MockStorage) GetOrdersByUserID(userID int) ([]domain.Order, error) {
	return m.GetOrdersByUserIDFn(userID)
}

func (m *MockStorage) GetUserByLogin(login string) (*domain.User, error) {
	return m.GetUserByLoginFn(login)
}

func (m *MockStorage) GetWithdrawalsByUserID(userID int) ([]domain.Withdrawal, error) {
	return m.GetWithdrawalsByUserIDFn(userID)
}

func (m *MockStorage) SaveUser(user domain.User) error {
	return m.SaveUserFn(user)
}

func (m *MockStorage) UpdateOrder(order domain.Order) error {
	return m.UpdateOrderFn(order)
}

func (m *MockStorage) UpdateUserBalance(userID int, amount float64) error {
	return m.UpdateUserBalanceFn(userID, amount)
}

func (m *MockStorage) LockOrderForProcessing(orderNumber string) error {
	return m.LockOrderForProcessingFn(orderNumber)
}

func (m *MockStorage) UnlockOrder(orderNumber string) error {
	return m.UnlockOrderFn(orderNumber)
}

func (m *MockStorage) GetUserBalance(login string) (float64, float64, error) {
	return m.GetUserBalanceFn(login)
}
