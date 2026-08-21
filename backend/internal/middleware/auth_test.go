package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProtectedRouteWithoutJWT(t *testing.T) {
	handler := Auth(fakeAuthenticator{}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", response.Code)
	}
}

type fakeAuthenticator struct{}

func (fakeAuthenticator) AuthenticateToken(token string) (string, error) {
	return "", nil
}
