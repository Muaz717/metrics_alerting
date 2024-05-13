package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func metricRouter() chi.Router {
	r := chi.NewRouter()

	r.Post("/update/{metricType}/{name}/{value}", handleWrongType)
	r.Post("/update/counter/{name}/{value}", handleCounter)
	r.Post("/update/gauge/{name}/{value}", handleGauge)
	r.Get("/value/{metricType}/{name}", giveValue)

	return r
}

func TestMetricsHandler(t *testing.T) {
	srv := httptest.NewServer(metricRouter())

	tests := []struct {
		url         string
		method      string
		contentType string
		code        int
		resp        string
	}{
		{"/update/counter/Counter1/11", http.MethodPost, "text/plain; charset=utf-8", http.StatusOK, ""},
		{"/update/gauge/Gauge1/21.1", http.MethodPost, "text/plain; charset=utf-8", http.StatusOK, ""},
		{"/update/counter/Counter1/12", http.MethodPost, "text/plain; charset=utf-8", http.StatusOK, ""},
		{"/value/counter/Counter1", http.MethodGet, "text/plain; charset=utf-8", http.StatusOK, "23"},
	}

	for _, test := range tests {
		req := resty.New().R()
		req.Method = test.method
		req.URL = srv.URL + test.url

		resp, err := req.Send()
		if err != nil {
			panic(err)
		}

		assert.Equal(t, test.code, resp.StatusCode())
		assert.Equal(t, test.contentType, resp.Header().Get("Content-Type"))

		assert.Equal(t, test.resp, string(resp.Body()))
	}
}
