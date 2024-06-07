package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/Muaz717/metrics_alerting/internal/logger"
	"github.com/Muaz717/metrics_alerting/internal/storag"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// Инициализация хранилища
var metricsStorage = &MemStorage{
	Gauges: make(map[string]float64),
	Counters: make(map[string]int64),
}

var mx = &sync.Mutex{}

func main() {
	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Use(logger.WithLogging)

		r.Get("/", GzipMiddleware(giveHTML))
		r.Post("/update/", GzipMiddleware(handleUpdateJSON))
		r.Post("/value/", GzipMiddleware(handleValueJSON))

		r.Post("/update/counter/{name}/{value}",handleCounter)
		r.Post("/update/gauge/{name}/{value}", handleGauge)
		r.Post("/update/{metricType}/{name}/{value}", handleWrongType)
		r.Get("/value/{metricType}/{name}", giveValue)
	})

	if err := logger.Initialize(flagLogLevel); err != nil{
		log.Fatal(err)
	}

	err := parseFlagsServer()
	if err != nil{
		log.Fatal(err)
	}

	fileName := flagFileStoragePath	

	Producer, err := NewProducer(fileName)
	if err != nil{
		log.Fatal(err)
	}
	defer Producer.Close()

	logger.Log.Info("Server is running on addr", zap.String("addr", flagRunAddr))
	log.Fatal(http.ListenAndServe(flagRunAddr, r))
}

func handleValueJSON(w http.ResponseWriter, r *http.Request){
	var metrics storage.Metrics

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&metrics); err != nil{
		logger.Log.Info("decoding request JSON body error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	switch metrics.MType{
	case "gauge":
		metricsStorage.ValueGaugeJSON(metrics, w)
	case "counter":
		metricsStorage.ValueCounterJSON(metrics, w)
	default:
		logger.Log.Info("Wrong metric type")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

}
func handleUpdateJSON(w http.ResponseWriter, r *http.Request) {
	var metrics storage.Metrics

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&metrics); err != nil{
		logger.Log.Info("decoding request JSON body error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	switch metrics.MType{
	case "gauge":
		metricsStorage.UpdateGaugeJSON(metrics, w)
	case "counter":
		metricsStorage.UpdateCounterJSON(metrics, w)
	default:
		logger.Log.Info("Wrong metric type")
		w.WriteHeader(http.StatusBadRequest)
		return
	}


}

func giveHTML(w http.ResponseWriter, r *http.Request){
	mx.Lock()
	defer mx.Unlock()
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	for name, value := range metricsStorage.Counters{
		wr := fmt.Sprintf("%s: %d\n", name, value)
		w.Write([]byte(wr))
	}

	for name, value := range metricsStorage.Gauges{
		wr1 := fmt.Sprintf("%s: %f\n", name, value)
		w.Write([]byte(wr1))
	}

}

func handleWrongType(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")

	if metricType != "gauge" && metricType != "counter"{
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func handleCounter(w http.ResponseWriter, r *http.Request){
	name := chi.URLParam(r, "name")
	value := chi.URLParam(r, "value")

	valueInt, err := strconv.ParseInt(value, 10, 64)
	if err != nil{
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	metricsStorage.UpdateCounter(name, valueInt)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

func handleGauge(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	value := chi.URLParam(r, "value")

	valueFloat, err := strconv.ParseFloat(value, 64)
	if err != nil{
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	metricsStorage.UpdateGauge(name, valueFloat)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

}

func giveValue(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")
	name := chi.URLParam(r, "name")

	switch metricType{
	case "counter":
		metricsStorage.GetCounter(name, w)
		return
	case "gauge":
		metricsStorage.GetGauge(name, w)
		return
	default:
		w.WriteHeader(http.StatusNotFound)
		return
	}
}
