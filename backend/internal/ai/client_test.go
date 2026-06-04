package ai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testChatReq() ChatRequest {
	return ChatRequest{Message: "привіт", UserID: 1}
}

func TestClient_Success(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"response":"ок","intent":"GENERAL"}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "secret")
	res, err := c.Chat(context.Background(), testChatReq())

	require.NoError(t, err)
	assert.Equal(t, "ок", res.Response)
	assert.Equal(t, "GENERAL", res.Intent)
	assert.Equal(t, int32(1), atomic.LoadInt32(&calls), "успіх — рівно один виклик")
}

func TestClient_SetsInternalTokenHeader(t *testing.T) {
	var gotToken, gotCT string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotToken = r.Header.Get("X-Internal-Token")
		gotCT = r.Header.Get("Content-Type")
		_, _ = w.Write([]byte(`{"response":"x","intent":"GENERAL"}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "my-shared-secret")
	_, err := c.Chat(context.Background(), testChatReq())

	require.NoError(t, err)
	assert.Equal(t, "my-shared-secret", gotToken)
	assert.Equal(t, "application/json", gotCT)
}

func TestClient_RetriesOn5xxThenSucceeds(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"response":"нарешті","intent":"GENERAL"}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "secret")
	res, err := c.Chat(context.Background(), testChatReq())

	require.NoError(t, err)
	assert.Equal(t, "нарешті", res.Response)
	assert.Equal(t, int32(3), atomic.LoadInt32(&calls), "дві невдачі + успіх = 3 спроби")
}

func TestClient_ExhaustsRetriesOn5xx(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "secret")
	_, err := c.Chat(context.Background(), testChatReq())

	require.Error(t, err)
	assert.Equal(t, int32(3), atomic.LoadInt32(&calls), "максимум 3 спроби на 5xx")
}

func TestClient_NoRetryOn4xx(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusBadRequest) // 4xx
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "secret")
	_, err := c.Chat(context.Background(), testChatReq())

	require.Error(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&calls), "4xx не повторюється")
}

func TestClient_RespectsContextCancellation(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	// Скасування під час backoff.
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	c := NewClient(srv.URL, "secret")
	_, err := c.Chat(ctx, testChatReq())

	require.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Less(t, atomic.LoadInt32(&calls), int32(3), "скасування перериває цикл спроб")
}
