package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	idempotencysvc "github.com/Asheze1127/progress-checker/backend/service/idempotency"
)

// mockStore is a test implementation of idempotencysvc.Store.
type mockStore struct {
	keys     map[string]bool
	existErr error
	setErr   error
}

func newMockStore() *mockStore {
	return &mockStore{keys: make(map[string]bool)}
}

func (m *mockStore) Exists(_ context.Context, key string) (bool, error) {
	if m.existErr != nil {
		return false, m.existErr
	}
	return m.keys[key], nil
}

func (m *mockStore) Set(_ context.Context, key string, _ time.Duration) error {
	if m.setErr != nil {
		return m.setErr
	}
	m.keys[key] = true
	return nil
}

func TestIdempotency(t *testing.T) {
	tests := []struct {
		name             string
		keyFunc          IdempotencyKeyFunc
		store            *mockStore
		retryHeader      string
		wantStatusCode   int
		wantHandlerCalls int32
	}{
		{
			name:             "new request passes through to handler",
			keyFunc:          func(_ *http.Request) string { return "unique-key" },
			store:            newMockStore(),
			wantStatusCode:   http.StatusOK,
			wantHandlerCalls: 1,
		},
		{
			name:    "duplicate request returns 200 without calling handler",
			keyFunc: func(_ *http.Request) string { return "duplicate-key" },
			store: func() *mockStore {
				s := newMockStore()
				s.keys["duplicate-key"] = true
				return s
			}(),
			wantStatusCode:   http.StatusOK,
			wantHandlerCalls: 0,
		},
		{
			name:             "empty key skips idempotency check and passes through",
			keyFunc:          func(_ *http.Request) string { return "" },
			store:            newMockStore(),
			wantStatusCode:   http.StatusOK,
			wantHandlerCalls: 1,
		},
		{
			name:    "store Exists error passes through gracefully",
			keyFunc: func(_ *http.Request) string { return "error-key" },
			store: func() *mockStore {
				s := newMockStore()
				s.existErr = errors.New("store unavailable")
				return s
			}(),
			wantStatusCode:   http.StatusOK,
			wantHandlerCalls: 1,
		},
		{
			name:    "store Set error still passes through to handler",
			keyFunc: func(_ *http.Request) string { return "set-error-key" },
			store: func() *mockStore {
				s := newMockStore()
				s.setErr = errors.New("write failure")
				return s
			}(),
			wantStatusCode:   http.StatusOK,
			wantHandlerCalls: 1,
		},
		{
			name:             "default key func with no retry header skips check",
			keyFunc:          nil,
			store:            newMockStore(),
			retryHeader:      "",
			wantStatusCode:   http.StatusOK,
			wantHandlerCalls: 1,
		},
		{
			name:             "default key func with retry header generates key",
			keyFunc:          nil,
			store:            newMockStore(),
			retryHeader:      "1",
			wantStatusCode:   http.StatusOK,
			wantHandlerCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var handlerCalls atomic.Int32

			handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				handlerCalls.Add(1)
				w.WriteHeader(http.StatusOK)
			})

			svc := idempotencysvc.NewService(tt.store)
			mw := Idempotency(svc, tt.keyFunc)
			wrapped := mw(handler)

			req := httptest.NewRequest(http.MethodPost, "/webhook/slack", nil)
			if tt.retryHeader != "" {
				req.Header.Set(headerSlackRetryNum, tt.retryHeader)
			}
			rec := httptest.NewRecorder()

			wrapped.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatusCode {
				t.Errorf("expected status %d, got %d", tt.wantStatusCode, rec.Code)
			}
			if got := handlerCalls.Load(); got != tt.wantHandlerCalls {
				t.Errorf("expected handler to be called %d times, got %d", tt.wantHandlerCalls, got)
			}
		})
	}
}
