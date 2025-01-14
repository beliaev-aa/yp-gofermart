package user

import (
	gofermartErrors "beliaev-aa/yp-gofermart/internal/gofermart/errors"
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
	validOrderNumber := "79927398713"
	mockOrderService := mocks.NewMockOrderServiceInterface(ctrl)
	mockUsernameExtractor := mocks.NewMockUsernameExtractor(ctrl)

	logger := zap.NewNop()
	handler := NewOrdersPostHandler(mockOrderService, mockUsernameExtractor, logger)

	testCases := []struct {
		name                 string
		requestBody          string
		setupMocks           func()
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:        "ExtractUsername_Error",
			requestBody: "123456",
			setupMocks: func() {
				mockUsernameExtractor.EXPECT().ExtractUsernameFromContext(gomock.Any(), gomock.Any()).Return("", errors.New("extraction error"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: "Internal Server Error\n",
		},
		{
			name:        "Invalid_Request_Body",
			requestBody: "",
			setupMocks: func() {
				mockUsernameExtractor.EXPECT().ExtractUsernameFromContext(gomock.Any(), gomock.Any()).Return("user", nil)
			},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "Invalid request format\n",
		},
		{
			name:        "Invalid_Order_Number_Format",
			requestBody: "invalid_order",
			setupMocks: func() {
				mockUsernameExtractor.EXPECT().ExtractUsernameFromContext(gomock.Any(), gomock.Any()).Return("user", nil)
			},
			expectedStatusCode:   http.StatusUnprocessableEntity,
			expectedResponseBody: "Invalid order number format\n",
		},
		{
			name:        "Order_Already_Uploaded",
			requestBody: validOrderNumber,
			setupMocks: func() {
				mockUsernameExtractor.EXPECT().ExtractUsernameFromContext(gomock.Any(), gomock.Any()).Return("user", nil)
				mockOrderService.EXPECT().AddOrder("user", validOrderNumber).Return(gofermartErrors.ErrOrderAlreadyUploaded)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: "",
		},
		{
			name:        "Order_Uploaded_By_Another",
			requestBody: validOrderNumber,
			setupMocks: func() {
				mockUsernameExtractor.EXPECT().ExtractUsernameFromContext(gomock.Any(), gomock.Any()).Return("user", nil)
				mockOrderService.EXPECT().AddOrder("user", validOrderNumber).Return(gofermartErrors.ErrOrderUploadedByAnother)
			},
			expectedStatusCode:   http.StatusConflict,
			expectedResponseBody: "Order number already uploaded by another user\n",
		},
		{
			name:        "Internal_Service_Error",
			requestBody: validOrderNumber,
			setupMocks: func() {
				mockUsernameExtractor.EXPECT().ExtractUsernameFromContext(gomock.Any(), gomock.Any()).Return("user", nil)
				mockOrderService.EXPECT().AddOrder("user", validOrderNumber).Return(errors.New("internal error"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: "Internal Server Error\n",
		},
		{
			name:        "Successful_Order_Addition",
			requestBody: validOrderNumber,
			setupMocks: func() {
				mockUsernameExtractor.EXPECT().ExtractUsernameFromContext(gomock.Any(), gomock.Any()).Return("user", nil)
				mockOrderService.EXPECT().AddOrder("user", validOrderNumber).Return(nil)
			},
			expectedStatusCode:   http.StatusAccepted,
			expectedResponseBody: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()

			req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(tc.requestBody))
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tc.expectedStatusCode {
				t.Errorf("Expected status code %d, got %d", tc.expectedStatusCode, rec.Code)
			}

			if rec.Body.String() != tc.expectedResponseBody {
				t.Errorf("Expected response body '%s', got '%s'", tc.expectedResponseBody, rec.Body.String())
			}
		})
	}
}
