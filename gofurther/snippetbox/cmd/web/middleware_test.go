package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ByChanderZap/snippetbox/internal/assert"
)

func TestCommonHeaders(t *testing.T) {
	rr := httptest.NewRecorder()

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	commonHeader(next).ServeHTTP(rr, req)

	res := rr.Result()
	defer res.Body.Close()

	expectedValue := "default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com"
	assert.Equal(t, expectedValue, res.Header.Get("Content-Security-Policy"))

	expectedValue = "origin-when-cross-origin"
	assert.Equal(t, expectedValue, res.Header.Get("Referrer-Policy"))

	expectedValue = "nosniff"
	assert.Equal(t, expectedValue, res.Header.Get("X-Content-Type-Options"))

	expectedValue = "deny"
	assert.Equal(t, expectedValue, res.Header.Get("X-Frame-Options"))

	expectedValue = "0"
	assert.Equal(t, expectedValue, res.Header.Get("X-XSS-Protection"))

	expectedValue = "Go"
	assert.Equal(t, expectedValue, res.Header.Get("Server"))

	assert.Equal(t, http.StatusOK, res.StatusCode)

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "OK", string(body))
}
