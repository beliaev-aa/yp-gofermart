package workers

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/services"
	"context"
	"go.uber.org/zap"
	"sync"
	"time"
)

// StartOrderStatusUpdater - запуск фонового процесса для обновления статусов заказов
func StartOrderStatusUpdater(ctx context.Context, orderService services.OrderServiceInterface, logger *zap.Logger, wg *sync.WaitGroup) {
	defer wg.Done()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var mu sync.Mutex
	isUpdating := false

	for {
		select {
		case <-ticker.C:
			mu.Lock()
			if !isUpdating {
				isUpdating = true
				mu.Unlock()

				orderCtx, cancel := context.WithTimeout(ctx, 4*time.Second)

				go func() {
					defer func() {
						mu.Lock()
						isUpdating = false
						mu.Unlock()
					}()
					defer cancel()

					orderService.UpdateOrderStatuses(orderCtx)
				}()
			} else {
				mu.Unlock()
				logger.Warn("Skipping order update due to ongoing processing")
			}
		case <-ctx.Done():
			logger.Info("Waiting for ongoing order status updates to complete")
			mu.Lock()
			for isUpdating {
				mu.Unlock()
				time.Sleep(100 * time.Millisecond)
				mu.Lock()
			}
			mu.Unlock()
			logger.Info("Shutting down order status update goroutine")
			return
		}
	}
}
