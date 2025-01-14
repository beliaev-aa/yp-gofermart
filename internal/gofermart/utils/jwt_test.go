package utils

import (
	"errors"
	"github.com/go-chi/jwtauth/v5"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"testing"
)

type testCase struct {
	Name             string
	ContextClaims    map[string]interface{}
	ExpectedUsername string
	ExpectedError    error
}

func TestRealUsernameExtractor_ExtractUsernameFromContext(t *testing.T) {
	testCases := []testCase{
		{
			Name: "Valid_Token_With_Username",
			ContextClaims: map[string]interface{}{
				"username": "test_user",
			},
			ExpectedUsername: "test_user",
			ExpectedError:    nil,
		},
		{
			Name: "Token_Without_Username",
			ContextClaims: map[string]interface{}{
				"email": "test@example.com",
			},
			ExpectedUsername: "",
			ExpectedError:    errors.New("unauthorized"),
		},
		{
			Name:             "Empty_Token_Claims",
			ContextClaims:    map[string]interface{}{},
			ExpectedUsername: "",
			ExpectedError:    errors.New("unauthorized"),
		},
	}

	extractor := &RealUsernameExtractor{}
	logger := zap.NewNop()

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)

			_, tokenString, _ := tokenAuth.Encode(tc.ContextClaims)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Authorization", "BEARER "+tokenString)

			token, err := jwtauth.VerifyRequest(tokenAuth, req, jwtauth.TokenFromHeader)
			if err != nil {
				t.Fatalf("failed to verify request: %v", err)
			}

			ctx := jwtauth.NewContext(req.Context(), token, nil)
			req = req.WithContext(ctx)

			username, err := extractor.ExtractUsernameFromContext(req, logger)

			if username != tc.ExpectedUsername {
				t.Errorf("expected username %v, got %v", tc.ExpectedUsername, username)
			}

			if err != nil && tc.ExpectedError == nil {
				t.Errorf("expected no error, but got %v", err)
			}

			if err == nil && tc.ExpectedError != nil {
				t.Errorf("expected error %v, but got no error", tc.ExpectedError)
			}

			if err != nil && tc.ExpectedError != nil && err.Error() != tc.ExpectedError.Error() {
				t.Errorf("expected error %v, got %v", tc.ExpectedError, err)
			}
		})
	}
}
