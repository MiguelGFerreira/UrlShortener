package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestMiddlewareAllowsBurstThenBlocks(t *testing.T) {
	h := New(0, 2).Middleware(okHandler()) // burst of 2, no refill

	want := []int{http.StatusOK, http.StatusOK, http.StatusTooManyRequests}
	for i, code := range want {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/shorten", nil)
		req.RemoteAddr = "10.0.0.1:1111"
		h.ServeHTTP(rec, req)
		if rec.Code != code {
			t.Errorf("request %d: code = %d, want %d", i+1, rec.Code, code)
		}
	}
}

func TestMiddlewareSeparatesClients(t *testing.T) {
	h := New(0, 1).Middleware(okHandler()) // burst of 1, no refill

	serve := func(ip string) int {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/shorten", nil)
		req.RemoteAddr = ip + ":2222"
		h.ServeHTTP(rec, req)
		return rec.Code
	}

	if got := serve("10.0.0.1"); got != http.StatusOK {
		t.Errorf("client A first: code = %d, want 200", got)
	}
	if got := serve("10.0.0.1"); got != http.StatusTooManyRequests {
		t.Errorf("client A second: code = %d, want 429", got)
	}
	if got := serve("10.0.0.2"); got != http.StatusOK {
		t.Errorf("client B first: code = %d, want 200", got)
	}
}
