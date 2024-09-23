package httpserver

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/services"
	"beliaev-aa/yp-gofermart/tests"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRegisterRoutes(t *testing.T) {
	logger := zap.NewNop()
	storageMock := &tests.MockStorage{}
	accrualMock := &tests.AccrualServiceMock{}

	appServices := &services.AppServices{
		UserService:  services.NewUserService(storageMock, logger),
		AuthService:  services.NewAuthService([]byte("secret"), logger, storageMock),
		OrderService: services.NewOrderService(accrualMock, storageMock, logger),
	}

	r := chi.NewRouter()
	RegisterRoutes(r, appServices, logger)

	testCases := []struct {
		name           string
		method         string
		path           string
		body           string
		expectedCode   int
		expectedBody   string
		expectedHeader string
	}{
		{
			name:         "Post_Api_User_Register",
			method:       http.MethodPost,
			path:         "/api/user/register",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Post_Api_User_Login",
			method:       http.MethodPost,
			path:         "/api/user/login",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Post_Api_User_Orders",
			method:       http.MethodPost,
			path:         "/api/user/orders",
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "Get_Api_User_Orders",
			method:       http.MethodGet,
			path:         "/api/user/orders",
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "Get_Api_User_Balance",
			method:       http.MethodGet,
			path:         "/api/user/balance",
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "Post_Api_User_Balance_Withdraw",
			method:       http.MethodPost,
			path:         "/api/user/balance/withdraw",
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "Get_Api_User_Withdrawals",
			method:       http.MethodGet,
			path:         "/api/user/withdrawals",
			expectedCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			if err != nil {
				t.Fatalf("could not create request: %v", err)
			}

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			if rr.Code != tt.expectedCode {
				t.Errorf("Unexpected code for %s: got %v, want %v", tt.name, rr.Code, tt.expectedCode)
			}
			if tt.expectedBody != "" && rr.Body.String() != tt.expectedBody {
				t.Errorf("Unexpected body for %s: got %v, want %v", tt.name, rr.Body.String(), tt.expectedBody)
			}
			if tt.expectedHeader != "" && rr.Header().Get("Location") != tt.expectedHeader {
				t.Errorf("Unexpected header(Location) for %s: got %v, want %v", tt.name, rr.Header().Get("Location"), tt.expectedHeader)
			}

			if tt.path == "/api/shorten" && tt.expectedCode == http.StatusCreated {
				var resp map[string]string
				err = json.NewDecoder(rr.Body).Decode(&resp)
				if err != nil {
					t.Fatalf("could not decode response: %v", err)
				}
				if _, exists := resp["result"]; !exists {
					t.Errorf("Expected 'result' key in response, but got: %v", resp)
				}
			}

			if tt.path == "/api/shorten/batch" && tt.expectedCode == http.StatusCreated {
				var responses []map[string]string
				err = json.NewDecoder(rr.Body).Decode(&responses)
				if err != nil {
					t.Fatalf("could not decode response: %v", err)
				}
				if len(responses) != 2 {
					t.Errorf("Expected 2 responses, got %d", len(responses))
				}
				if responses[0]["correlation_id"] != "123" || responses[1]["correlation_id"] != "124" {
					t.Errorf("Unexpected correlation ids: %v, %v", responses[0]["correlation_id"], responses[1]["correlation_id"])
				}
			}
		})
	}
}
