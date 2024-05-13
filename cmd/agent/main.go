package main

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"time"
)

const (
    pollInterval   = 2
    reportInterval = 10
    serverAddress  = "http://localhost:8080"
)



func updateMetrics(metrics *map[string]float64) {
	metric := new(runtime.MemStats)
	runtime.ReadMemStats(metric)


	for {
		pollCount++
		randomValue = rand.Float64()

		// Сбор метрик из пакета runtime
		*metrics = map[string]float64{
            "Alloc":        float64(metric.Alloc),
            "BuckHashSys":  float64(metric.BuckHashSys),
			"Frees": 		float64(metric.Frees),
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
			"Mallocs": 		float64(metric.Mallocs),
			"NextGC": 		float64(metric.NextGC),
			"NumForcedGC": 	float64(metric.NumForcedGC),
			"NumGC": 		float64(metric.NumGC),
			"OtherSys": 	float64(metric.OtherSys),
			"PauseTotalNs": float64(metric.PauseTotalNs),
			"StackInuse": 	float64(metric.StackInuse),
			"StackSys": 	float64(metric.StackSys),
			"Sys": 			float64(metric.Sys),
			"TotalAlloc": 	float64(metric.TotalAlloc),
            "RandomValue":  randomValue,
        }

		time.Sleep(pollInterval * time.Second)
	}
}
func sendMetric(metrics map[string]float64, pollCount int64) {

	for metricName, value := range metrics{

		url := fmt.Sprintf("%s/update/gauge/%s/%f", serverAddress, metricName, value)

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(nil))
		if err != nil{
			log.Println("Error sending metric:", err)
			return
		}
		req.Header.Set("Content-Type", "text/plain")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
            log.Println("Error sending metric:", err)
            return
        }
		log.Println(resp)
		defer resp.Body.Close()


	}

	url1 := fmt.Sprintf("%s/update/counter/PollCount/%d", serverAddress, pollCount)

	req, err := http.NewRequest(http.MethodPost, url1, bytes.NewBuffer(nil))
	if err != nil{
		log.Println("Error sending metric:", err)
		return
	}
	req.Header.Set("Content-Type", "text/plain")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
        log.Println("Error sending metric:", err)
        return
    }
	log.Println(resp)
	defer resp.Body.Close()
}
	var pollCount int64
    var randomValue = 0.0

func main() {
	metrics := map[string]float64{}

	go updateMetrics(&metrics)

	for{
		go sendMetric(metrics, pollCount)
		time.Sleep(reportInterval * time.Second)
	}

}
