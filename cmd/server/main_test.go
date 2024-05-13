package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetricsHandler(t *testing.T) {
	type Want struct{
		code int
		contentType string
		contentLength int
	}

	// type request struct{

	// }

	tests := []struct{
		name string
		want Want
	}{
		{
			name: "simple test#1",
			want: Want{
				code: 200,
				contentType: "text/plain; charset=utf-8",
				contentLength: 0,
			},
		},
	}

	for _, test := range tests{
		t.Run(test.name, func(t *testing.T){
			request := httptest.NewRequest(http.MethodPost, "http://localhost:8080/update/counter/someMetric/432", nil)

			w := httptest.NewRecorder()
			handleMetric(w, request)
			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, test.want.code, res.StatusCode)
		})
	}
}
