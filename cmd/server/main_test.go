package main

import (
	"net/http"
	// "net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	// "github.com/go-resty/resty/v2"
	// "github.com/stretchr/testify/assert"
)

func metricRouter() chi.Router {
	r := chi.NewRouter()

	r.Post("/update/{metricType}/{name}/{value}", handleMetric)
	r.Get("/value/{metricType}/{name}", getValue)

	return r
}

func TestMetricsHandler(t *testing.T) {
	// srv := httptest.NewServer(metricRouter())

	tests := []struct {
		name		string
		url         string
		method      string
		contentType string
		code        int
		resp        string
	}{
		{"1", "/update/counter/Counter1/11", http.MethodPost, "text/plain; charset=utf-8", http.StatusOK, ""},
		{"2", "/update/gauge/Gauge1/21.1", http.MethodPost, "text/plain; charset=utf-8", http.StatusOK, ""},
		{"3", "/update/counter/Counter1/12", http.MethodPost, "text/plain; charset=utf-8", http.StatusOK, ""},
		{"4", "/value/counter/Counter1", http.MethodGet, "text/plain; charset=utf-8", http.StatusOK, "23"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// r := httptest.NewRequest(tt.method, tt.url, http.NoBody)
			// w := httptest.NewRecorder()
			// tt.s.UpdateHandler(w, r)
		})
	}
}
