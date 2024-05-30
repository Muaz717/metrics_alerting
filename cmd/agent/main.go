package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"sync"
	"time"

	storage "github.com/Muaz717/metrics_alerting/internal/storag"
	"github.com/go-resty/resty/v2"
)

const (
    pollInterval   = 2
    reportInterval = 10
    serverAddress  = "http://"
)

var pollCount int64
var randomValue = 0.0
var metric = new(runtime.MemStats)


var mx = &sync.Mutex{}

func updateMetrics(metrics *map[string]interface{}) {
	runtime.ReadMemStats(metric)

	mx.Lock()
	randomValue = rand.Float64()
	pollCount++
		// Сбор метрик из пакета runtime
	*metrics = map[string]interface{}{
        "Alloc":        float64(metric.Alloc),
        "BuckHashSys":  float64(metric.BuckHashSys),
		"GCCPUFraction":float64(metric.GCCPUFraction),
		"GCSys": 		float64(metric.GCSys),
		"HeapAlloc": 	float64(metric.HeapAlloc),
		"HeapIdle": 	float64(metric.HeapIdle),
		"HeapInuse": 	float64(metric.HeapInuse),
		"HeapObjects": 	float64(metric.HeapObjects),
		"HeapReleased":	float64(metric.HeapReleased),
		"HeapSys": 		float64(metric.HeapSys),
		"LastGC": 		float64(metric.LastGC),
		"Lookups": 		float64(metric.Lookups),
		"MCacheInuse": 	float64(metric.MCacheInuse),
		"MCacheSys": 	float64(metric.MCacheSys),
		"MSpanInuse": 	float64(metric.MSpanInuse),
		"MSpanSys":		float64(metric.MSpanSys),
		"NextGC": 		float64(metric.NextGC),
		"NumForcedGC": 	float64(metric.NumForcedGC),
		"Mallocs": 		float64(metric.Mallocs),
		"NumGC": 		float64(metric.NumGC),
		"OtherSys": 	float64(metric.OtherSys),
		"PauseTotalNs": float64(metric.PauseTotalNs),
		"StackInuse": 	float64(metric.StackInuse),
		"StackSys": 	float64(metric.StackSys),
		"Sys": 			float64(metric.Sys),
		"TotalAlloc": 	float64(metric.TotalAlloc),
        "RandomValue":  randomValue,
		"PollCount":	pollCount,
		"Frees": 		float64(metric.Frees),
	}
	log.Println("Metrics updated")
	mx.Unlock()
}

func sendMetric(metrics map[string]interface{}) {
	mx.Lock()
	defer mx.Unlock()
	time.Sleep(1 * time.Second)
	var client = resty.New()

	for metricName, value := range metrics{

		switch value.(type){
		case float64:
			url := fmt.Sprintf("%s/update/gauge/%s/%f", serverAddress+flags.flagRunAddr, metricName, value)

		_, err := client.R().
			SetHeader("Content-Type", "text/plain").
			Post(url)

		if err != nil{
			log.Println("Error sending metric:", err)
			return
		}

	case int64:
		url := fmt.Sprintf("%s/update/counter/%s/%d", serverAddress+flags.flagRunAddr, metricName, value)

		_, err := client.R().
			SetHeader("Content-Type", "text/plain").
			Post(url)

		if err != nil{
			log.Println("Error sending metric:", err)
			return
		}
	}
		}

}

func sendMetricJSON(metrics map[string]interface{}){
	var metricsJSON storage.Metrics
	url := fmt.Sprintf("%s/update/", serverAddress+flags.flagRunAddr)

	for metricName, value := range metrics{
		mx.Lock()
		switch val := value.(type){
		case float64:
			metricsJSON = storage.Metrics{
				ID: metricName,
				MType: "gauge",
				Value: &val,
			}

		case int64:
			metricsJSON = storage.Metrics{
				ID: metricName,
				MType: "counter",
				Delta: &val,
			}

		}
		mx.Unlock()

		var buffer bytes.Buffer
		err := json.NewEncoder(&buffer).Encode(metricsJSON)
		if err != nil {
			log.Printf("failed to JSON encode gauge metric: %v", err)
			return
		}

		client := resty.New()
		_, err = client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(metricsJSON).
			Post(url)

		if err  != nil{
			log.Println("Error sending JSON metric:", err)
			return
		}

		log.Printf("Metric %v has been sent", metricName)
	}

}

func main() {
	metrics := map[string]interface{}{}

	err := parseFlagsAgent()
	if err != nil{
		log.Fatal(err)
	}

	tickerUpdate := time.NewTicker(time.Duration(flags.flagPollInterval) * time.Second)
	tickerSendJSON := time.NewTicker(time.Duration(flags.flagReportInterval) * time.Second)

	for {
		select{
		case <-tickerUpdate.C:
			updateMetrics(&metrics)
		case <-tickerSendJSON.C:
			sendMetricJSON(metrics)
		}
	}

}
