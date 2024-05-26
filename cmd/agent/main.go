package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
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

var mx = &sync.Mutex{}

func updateMetrics(metrics *map[string]interface{}) {
	metric := new(runtime.MemStats)
	runtime.ReadMemStats(metric)


	for {
		mx.Lock()
		pollCount++
		randomValue = rand.Float64()

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
		mx.Unlock()

		time.Sleep(time.Duration(flags.flagPollInterval) * time.Second)
		// log.Println(metric.Frees)
		// log.Println(metric.Alloc)
		// log.Println(metric.Mallocs)
		// log.Println(metric.Sys)
	}
}

func sendMetric(metrics map[string]interface{}) {
	mx.Lock()
	defer mx.Unlock()
	var client = resty.New()
	client.SetTimeout(1 * time.Second)

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

		// log.Printf("Metric %v has been sent", metricName)

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

	client := resty.New()

	for metricName, value := range metrics{

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

		var buffer bytes.Buffer
		err := json.NewEncoder(&buffer).Encode(metricsJSON)
		if err != nil {
			log.Printf("failed to JSON encode gauge metric: %v", err)
			return
		}

		req := client.R()
		req.Method = http.MethodPost
		req.Body = metricsJSON
		req.SetHeader("Content-Type", "application/json")
		req.URL = url

		_, err = req.Send()
		if err  != nil{
			log.Println("Error sending JSON metric:", err)
			return
		}

	}

}

func main() {

	metrics := map[string]interface{}{}

	err := parseFlagsAgent()
	if err != nil{
		log.Fatal(err)
	}


	go updateMetrics(&metrics)

	for{
		go sendMetric(metrics)
		go sendMetricJSON(metrics)
		time.Sleep(time.Duration(flags.flagReportInterval) * time.Second)
	}

}
