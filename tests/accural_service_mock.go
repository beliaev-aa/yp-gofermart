package tests

import "context"

// AccrualServiceMock - это мок для внешнего сервиса начислений
type AccrualServiceMock struct {
	GetOrderAccrualFn func(orderNumber string) (float64, string, error)
}

func (m *AccrualServiceMock) GetOrderAccrual(_ context.Context, orderNumber string) (float64, string, error) {
	return m.GetOrderAccrualFn(orderNumber)
}
