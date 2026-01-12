package cache

import (
	"context"
	"time"
)

// MockClient is a mock implementation of Client interface for testing
type MockClient struct {
	// Key-value store
	store map[string]mockValue

	// Pub/sub
	subscribers map[string][]func(channel, message string)

	// Error injection for testing
	SetError       error
	GetError       error
	DeleteError    error
	ExistsError    error
	TTLError       error
	ExpireError    error
	PublishError   error
	SubscribeError error
	PingError      error
	CloseError     error
}

type mockValue struct {
	value string
	ttl   time.Duration
	setAt time.Time
}

// NewMockClient creates a new mock cache client for testing
func NewMockClient() *MockClient {
	return &MockClient{
		store:       make(map[string]mockValue),
		subscribers: make(map[string][]func(channel, message string)),
	}
}

func (m *MockClient) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if m.SetError != nil {
		return m.SetError
	}

	var strValue string
	switch v := value.(type) {
	case string:
		strValue = v
	default:
		strValue = "mocked"
	}

	m.store[key] = mockValue{
		value: strValue,
		ttl:   ttl,
		setAt: time.Now(),
	}
	return nil
}

func (m *MockClient) Get(ctx context.Context, key string) (string, error) {
	if m.GetError != nil {
		return "", m.GetError
	}

	val, exists := m.store[key]
	if !exists {
		return "", ErrKeyNotFound
	}

	// Check if TTL expired
	if val.ttl > 0 && time.Since(val.setAt) > val.ttl {
		delete(m.store, key)
		return "", ErrKeyNotFound
	}

	return val.value, nil
}

func (m *MockClient) Delete(ctx context.Context, keys ...string) error {
	if m.DeleteError != nil {
		return m.DeleteError
	}

	for _, key := range keys {
		delete(m.store, key)
	}
	return nil
}

func (m *MockClient) Exists(ctx context.Context, key string) (bool, error) {
	if m.ExistsError != nil {
		return false, m.ExistsError
	}

	_, exists := m.store[key]
	return exists, nil
}

func (m *MockClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	if m.TTLError != nil {
		return 0, m.TTLError
	}

	val, exists := m.store[key]
	if !exists {
		return -2 * time.Second, nil
	}

	if val.ttl == 0 {
		return -1, nil
	}

	elapsed := time.Since(val.setAt)
	remaining := val.ttl - elapsed
	if remaining < 0 {
		return -2 * time.Second, nil
	}

	return remaining, nil
}

func (m *MockClient) Expire(ctx context.Context, key string, ttl time.Duration) error {
	if m.ExpireError != nil {
		return m.ExpireError
	}

	val, exists := m.store[key]
	if !exists {
		return nil
	}

	val.ttl = ttl
	val.setAt = time.Now()
	m.store[key] = val
	return nil
}

func (m *MockClient) Publish(ctx context.Context, channel string, message interface{}) error {
	if m.PublishError != nil {
		return m.PublishError
	}

	callbacks, exists := m.subscribers[channel]
	if !exists {
		return nil
	}

	msgStr, ok := message.(string)
	if !ok {
		msgStr = "mocked_message"
	}

	for _, callback := range callbacks {
		callback(channel, msgStr)
	}

	return nil
}

func (m *MockClient) Subscribe(ctx context.Context, callback func(channel, message string), channels ...string) (cancel func() error, err error) {
	if m.SubscribeError != nil {
		return nil, m.SubscribeError
	}

	if len(channels) == 0 {
		return nil, ErrPubSubFailed
	}

	for _, channel := range channels {
		m.subscribers[channel] = append(m.subscribers[channel], callback)
	}

	cancelFunc := func() error {
		for _, channel := range channels {
			delete(m.subscribers, channel)
		}
		return nil
	}

	return cancelFunc, nil
}

func (m *MockClient) Ping(ctx context.Context) error {
	return m.PingError
}

func (m *MockClient) Close() error {
	return m.CloseError
}

// Reset clears all stored data and errors (useful for test cleanup)
func (m *MockClient) Reset() {
	m.store = make(map[string]mockValue)
	m.subscribers = make(map[string][]func(channel, message string))
	m.SetError = nil
	m.GetError = nil
	m.DeleteError = nil
	m.ExistsError = nil
	m.TTLError = nil
	m.ExpireError = nil
	m.PublishError = nil
	m.SubscribeError = nil
	m.PingError = nil
	m.CloseError = nil
}
