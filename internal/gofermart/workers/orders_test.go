package workers

import (
	"beliaev-aa/yp-gofermart/tests/mocks"
	"context"
	"github.com/golang/mock/gomock"
	"go.uber.org/zap"
	"sync"
	"testing"
	"time"
)

func TestStartOrderStatusUpdater(t *testing.T) {
	testCases := []struct {
		name           string
		mockUpdateFunc func(*mocks.MockOrderServiceInterface)
		expectedWait   bool
	}{
		{
			name: "Update_Successfully",
			mockUpdateFunc: func(m *mocks.MockOrderServiceInterface) {
				m.EXPECT().UpdateOrderStatuses(gomock.Any()).Times(1)
			},
			expectedWait: false,
		},
		{
			name: "Skip_Update_Due_To_Processing",
			mockUpdateFunc: func(m *mocks.MockOrderServiceInterface) {
				m.EXPECT().UpdateOrderStatuses(gomock.Any()).Do(func(ctx context.Context) {
					time.Sleep(1500 * time.Millisecond)
				}).Times(1)
			},
			expectedWait: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			MockOrderServiceInterface := mocks.NewMockOrderServiceInterface(ctrl)
			tc.mockUpdateFunc(MockOrderServiceInterface)

			wg := &sync.WaitGroup{}
			wg.Add(1)

			ctx, cancel := context.WithCancel(context.Background())

			go StartOrderStatusUpdater(ctx, MockOrderServiceInterface, zap.NewNop(), wg)

			if tc.expectedWait {
				time.Sleep(2000 * time.Millisecond)
			} else {
				time.Sleep(1200 * time.Millisecond)
			}

			cancel()
			wg.Wait()
		})
	}
}
