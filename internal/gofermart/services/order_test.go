package services

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	gofermartErrors "beliaev-aa/yp-gofermart/internal/gofermart/errors"
	"beliaev-aa/yp-gofermart/tests/mocks"
	"errors"
	"github.com/golang/mock/gomock"
	"go.uber.org/zap"
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
