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
var metricsStorage = &storage.MemStorage{
	Gauges: make(map[string]float64),
	Counters: make(map[string]int64),
}

var mx = &sync.Mutex{}

func main() {
	if err := logger.Initialize(flagLogLevel); err != nil{
		log.Fatal(err)
	}

	r := chi.NewRouter()

	r.Get("/", logger.WithLogging(giveHTML))
	r.Post("/update/", logger.WithLogging(handleSendJSON))
	r.Post("/value/", logger.WithLogging(handleGetJSON))
	r.Post("/update/counter/{name}/{value}", logger.WithLogging(handleCounter))
	r.Post("/update/gauge/{name}/{value}", logger.WithLogging(handleGauge))
	r.Post("/update/{metricType}/{name}/{value}", logger.WithLogging(handleWrongType))
	r.Get("/value/{metricType}/{name}", logger.WithLogging(giveValue))

	parseFlagsServer()

	logger.Log.Info("Server is running on addr", zap.String("addr", flagRunAddr))
	log.Fatal(http.ListenAndServe(flagRunAddr, r))
}

func handleGetJSON(w http.ResponseWriter, r *http.Request){
	var metrics storage.Metrics

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&metrics); err != nil{
		logger.Log.Info("decoding request JSON body error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}


	response := storage.Metrics{
		ID: metrics.ID,
		MType: metrics.MType,
	}
	switch response.MType{
	case "counter":
		mx.Lock()
		if _, ok := metricsStorage.Counters[response.ID]; !ok{
			logger.Log.Info("No counter metric with this id")
			w.WriteHeader(http.StatusNotFound)
			return
		}
		val := metricsStorage.Counters[response.ID]
		response.Delta = &val

		w.Header().Set("Content-type", "application/json")

		enc := json.NewEncoder(w)
		if err := enc.Encode(response); err != nil{
			logger.Log.Info("encoding response JSON body error", zap.Error(err))
			return
		}
		w.WriteHeader(http.StatusOK)
		mx.Unlock()
	case "gauge":
		mx.Lock()
		if _, ok := metricsStorage.Gauges[response.ID]; !ok{
			logger.Log.Info("No gauge metric with this id")
			w.WriteHeader(http.StatusNotFound)
			return
		}
		value := metricsStorage.Gauges[response.ID]
		response.Value = &value

		w.Header().Set("Content-type", "application/json")

		enc := json.NewEncoder(w)
		if err := enc.Encode(response); err != nil{
			logger.Log.Info("encoding response JSON body error", zap.Error(err))
			return
		}
		w.WriteHeader(http.StatusOK)
		mx.Unlock()
	}

}

func handleSendJSON(w http.ResponseWriter, r *http.Request) {
	var metrics storage.Metrics

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&metrics); err != nil{
		logger.Log.Info("decoding request JSON body error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	switch metrics.MType{
	case "gauge":
		mx.Lock()
		response := storage.Metrics{
			ID: metrics.ID,
			MType: metrics.MType,
			Value: metrics.Value,
		}

		if response.ID == ""{
			logger.Log.Info("Forgot metric name")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		metricsStorage.Gauges[response.ID] = *metrics.Value

		enc := json.NewEncoder(w)
		if err := enc.Encode(response); err != nil{
			logger.Log.Info("encoding response JSON body error", zap.Error(err))
			return
		}
		mx.Unlock()
	case "counter":
		mx.Lock()
		response := storage.Metrics{
			ID: metrics.ID,
			MType: metrics.MType,
			Delta: metrics.Delta,
		}

		if response.ID == ""{
			logger.Log.Info("Forgot metric name")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if value, ok := metricsStorage.Counters[response.ID]; ok{
			newValue := *response.Delta + value
			response.Delta = &newValue
		}
		metricsStorage.Counters[response.ID] = *response.Delta

		enc := json.NewEncoder(w)
		if err := enc.Encode(response); err != nil{
			logger.Log.Info("encoding response JSON body error", zap.Error(err))
			return
		}
		mx.Unlock()
	default:
		logger.Log.Info("Wrong metric type")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)

}

func giveHTML(w http.ResponseWriter, r *http.Request){
	for name, value := range metricsStorage.Counters{
		wr := fmt.Sprintf("%s: %d\n", name, value)
		w.Write([]byte(wr))
	}

	for name, value := range metricsStorage.Gauges{
		wr1 := fmt.Sprintf("%s: %f\n", name, value)
		w.Write([]byte(wr1))
	}
	w.WriteHeader(http.StatusOK)
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
