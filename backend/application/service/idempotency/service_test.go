package idempotency

import (
	"context"
	"errors"
	"testing"
	"time"
)

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

func TestService_IsDuplicate(t *testing.T) {
	tests := []struct {
		name string
		store *mockStore
		key   string
		want  bool
	}{
		{
			name:  "new key returns false and stores it",
			store: newMockStore(),
			key:   "new-key",
			want:  false,
		},
		{
			name: "existing key returns true",
			store: func() *mockStore {
				s := newMockStore()
				s.keys["existing-key"] = true
				return s
			}(),
			key:  "existing-key",
			want: true,
		},
		{
			name: "store Exists error returns false",
			store: func() *mockStore {
				s := newMockStore()
				s.existErr = errors.New("store unavailable")
				return s
			}(),
			key:  "error-key",
			want: false,
		},
		{
			name: "store Set error still returns false",
			store: func() *mockStore {
				s := newMockStore()
				s.setErr = errors.New("write failure")
				return s
			}(),
			key:  "set-error-key",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(tt.store)
			got := svc.IsDuplicate(context.Background(), tt.key)
			if got != tt.want {
				t.Errorf("IsDuplicate(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}

func TestService_IsDuplicate_stores_key_on_first_call(t *testing.T) {
	store := newMockStore()
	svc := NewService(store)

	if svc.IsDuplicate(context.Background(), "key") {
		t.Fatal("first call should return false")
	}
	if !store.keys["key"] {
		t.Fatal("key should be stored after first call")
	}
	if !svc.IsDuplicate(context.Background(), "key") {
		t.Fatal("second call should return true")
	}
}
