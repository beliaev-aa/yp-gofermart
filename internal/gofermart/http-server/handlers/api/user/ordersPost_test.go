package user

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	gofermartErrors "beliaev-aa/yp-gofermart/internal/gofermart/errors"
	"beliaev-aa/yp-gofermart/internal/gofermart/services"
	"beliaev-aa/yp-gofermart/tests"
	"beliaev-aa/yp-gofermart/tests/mocks"
	"errors"
	"github.com/golang/mock/gomock"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOrdersPostHandler_ServeHTTP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	logger := zap.NewNop()

	validOrderNumber := "79927398713"
	mockExtractor := mocks.NewMockUsernameExtractor(ctrl)

	testCases := []struct {
		name               string
		login              string
		body               string
		mockExtractFn      func(r *http.Request, logger *zap.Logger) (string, error)
		mockSetup          func(m *tests.MockStorage)
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name: "Unauthorized_Access",
			mockExtractFn: func(r *http.Request, logger *zap.Logger) (string, error) {
				return "", http.ErrNoCookie
			},
			body:               validOrderNumber,
			expectedStatusCode: http.StatusInternalServerError,
			expectedBody:       "Internal Server Error\n",
		},
		{
			name: "Invalid_Request_Body",
			mockExtractFn: func(r *http.Request, logger *zap.Logger) (string, error) {
				return "test_user", nil
			},
			body:               "",
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       "Invalid request format\n",
		},
		{
			name: "Invalid_Order_Number_Format",
			mockExtractFn: func(r *http.Request, logger *zap.Logger) (string, error) {
				return "test_user", nil
			},
			body:               "123456",
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedBody:       "Invalid order number format\n",
		},
		{
			name: "Order_Already_Uploaded_By_User",
			mockExtractFn: func(r *http.Request, logger *zap.Logger) (string, error) {
				return "test_user", nil
			},
			mockSetup: func(m *tests.MockStorage) {
				m.AddOrderFn = func(order domain.Order) error {
					return gofermartErrors.ErrOrderAlreadyUploaded
				}
				m.GetUserByLoginFn = func(login string) (*domain.User, error) {
					return &domain.User{UserID: 1}, nil
				}
				m.GetOrderByNumberFn = func(number string) (*domain.Order, error) {
					return &domain.Order{UserID: 2, OrderNumber: validOrderNumber}, nil
				}
			},
			body:               validOrderNumber,
			expectedStatusCode: http.StatusConflict,
			expectedBody:       "Order number already uploaded by another user\n",
		},
		{
			name: "Internal_Server_Error_On_Add_Order",
			mockExtractFn: func(r *http.Request, logger *zap.Logger) (string, error) {
				return "test_user", nil
			},
			mockSetup: func(m *tests.MockStorage) {
				m.AddOrderFn = func(order domain.Order) error {
					return errors.New("failed to add order")
				}
				m.GetUserByLoginFn = func(login string) (*domain.User, error) {
					return &domain.User{UserID: 1}, nil
				}
				m.GetOrderByNumberFn = func(number string) (*domain.Order, error) {
					return nil, errors.New("failed to add order")
				}
			},
			body:               validOrderNumber,
			expectedStatusCode: http.StatusInternalServerError,
			expectedBody:       "Internal Server Error\n",
		},
		{
			name: "Order_Accepted",
			mockExtractFn: func(r *http.Request, logger *zap.Logger) (string, error) {
				return "test_user", nil
			},
			mockSetup: func(m *tests.MockStorage) {
				m.AddOrderFn = func(order domain.Order) error {
					return nil
				}
				m.GetUserByLoginFn = func(login string) (*domain.User, error) {
					return &domain.User{UserID: 1}, nil
				}
				m.GetOrderByNumberFn = func(number string) (*domain.Order, error) {
					return nil, gofermartErrors.ErrOrderNotFound
				}
			},
			body:               validOrderNumber,
			expectedStatusCode: http.StatusAccepted,
			expectedBody:       "",
		},
		{
			name: "Order_Already_Uploaded_By_Same_User",
			mockExtractFn: func(r *http.Request, logger *zap.Logger) (string, error) {
				return "test_user", nil
			},
			mockSetup: func(m *tests.MockStorage) {
				m.AddOrderFn = func(order domain.Order) error {
					return gofermartErrors.ErrOrderAlreadyUploaded
				}
				m.GetUserByLoginFn = func(login string) (*domain.User, error) {
					return &domain.User{UserID: 1}, nil
				}
				m.GetOrderByNumberFn = func(number string) (*domain.Order, error) {
					return &domain.Order{UserID: 1, OrderNumber: validOrderNumber}, nil
				}
			},
			body:               validOrderNumber,
			expectedStatusCode: http.StatusOK,
			expectedBody:       "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockExtractor.EXPECT().ExtractUsernameFromContext(gomock.Any(), gomock.Any()).DoAndReturn(tc.mockExtractFn)

			mockStorage := &tests.MockStorage{}
			if tc.mockSetup != nil {
				tc.mockSetup(mockStorage)
			}

			orderService := services.NewOrderService(nil, mockStorage, logger)

			handler := NewOrdersPostHandler(orderService, mockExtractor, logger)

			req := httptest.NewRequest("POST", "/orders", strings.NewReader(tc.body))

			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tc.expectedStatusCode {
				t.Errorf("expected status %v, got %v", tc.expectedStatusCode, rr.Code)
			}

			if rr.Body.String() != tc.expectedBody {
				t.Errorf("expected body %q, got %q", tc.expectedBody, rr.Body.String())
			}
		})
	}
}
