package main

import (
	"bytes"
	"compress/gzip"

	// "encoding/json"
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/Muaz717/metrics_alerting/internal/config"
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

func sendMetric(metrics map[string]interface{}, cfg config.AgentCfg) {
	mx.Lock()
	defer mx.Unlock()
	time.Sleep(1 * time.Second)
	var client = resty.New()

	for metricName, value := range metrics{

		switch value.(type){
		case float64:
			url := fmt.Sprintf("%s/update/gauge/%s/%f", serverAddress+cfg.Host, metricName, value)

		_, err := client.R().
			SetHeader("Content-Type", "text/plain").
			Post(url)

		if err != nil{
			log.Println("Error sending metric:", err)
			return
		}

	case int64:
		url := fmt.Sprintf("%s/update/counter/%s/%d", serverAddress+cfg.Host, metricName, value)

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

func sendMetricJSON(metrics map[string]interface{},cfg config.AgentCfg){
	var metricsJSON storage.Metrics
	url := fmt.Sprintf("%s/update/", serverAddress+cfg.Host)

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


		// data, err := json.Marshal(metricsJSON)
		// if err != nil{
		// 	log.Println("Error serializing JSON to data")
		// 	return
		// }
		// compressedData, err := CompressData(data)
		// if err != nil{
		// 	log.Println("Compress error")
		// 	return
		// }

		client := resty.New()
		_, err := client.R().
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

func CompressData(data []byte) ([]byte, error){
	var buff bytes.Buffer

	gz := gzip.NewWriter(&buff)

	_, err := gz.Write(data)
	if err != nil{
		return nil, fmt.Errorf("failed write data to compress temporary buffer: %v", err)
	}

	err = gz.Close()
	if err != nil {
        return nil, fmt.Errorf("failed compress data: %v", err)
    }

	return buff.Bytes(), nil
}

func main() {
	metrics := map[string]interface{}{}

	cfg, err := config.NewAgentConfiguration()
	if err != nil{
		log.Fatal(err)
	}

	tickerUpdate := time.NewTicker(time.Duration(cfg.PollInterval) * time.Second)
	tickerSendJSON := time.NewTicker(time.Duration(cfg.ReportInterval) * time.Second)

	for {
		select{
		case <-tickerUpdate.C:
			updateMetrics(&metrics)
		case <-tickerSendJSON.C:
			sendMetricJSON(metrics, cfg)
		}
	}

}
