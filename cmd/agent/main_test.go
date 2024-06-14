package main

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/Muaz717/metrics_alerting/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestSendMetric(t *testing.T) {
    mx := &sync.Mutex{}
    metrics := map[string]interface{}{"RandomValue": 0.5, "PollCount": 1}

    cfg, err := config.NewAgentConfiguration()
	if err != nil{
		log.Fatal(err)
	}

    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        body, err := io.ReadAll(r.Body)
        assert.NoError(t, err)

        switch r.RequestURI {
        case "/update/gauge/RandomValue":
            assert.Equal(t, "0.500000", string(body))
        case "/update/counter/PollCount":
            assert.Equal(t, "1", string(body))
        }
    }))
    defer ts.Close()

    go sendMetric(metrics, cfg)
    time.Sleep(time.Second)

    mx.Lock()
    assert.True(t, metrics["RandomValue"] == 0.5)
    assert.True(t, metrics["PollCount"] == 1)
    mx.Unlock()
}

func TestUpdateMetrics(t *testing.T) {
    mx := &sync.Mutex{}
    metrics := map[string]interface{}{}

    go updateMetrics(&metrics)
    time.Sleep(time.Second)

    mx.Lock()
    assert.NotEqual(t, 0.0, metrics["RandomValue"])
    assert.True(t, pollCount > 0)
    mx.Unlock()
}
