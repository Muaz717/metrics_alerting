package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/Muaz717/metrics_alerting/internal/logger"
	"github.com/Muaz717/metrics_alerting/internal/models"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// Тип хранилища для метрик
type MemStorage struct {
	mx sync.Mutex
	Gauges map[string]float64
	Counters map[string]int64
}

// Интерфейс для взаимодействия с хранилищем
type Storage interface {
	GetGauge(string, http.ResponseWriter)
	GetCounter(string, http.ResponseWriter)
	UpdateGauge(string, float64)
	UpdateCounter(string, int64)
}

// Инициализация хранилища
var storage = &MemStorage{
	Gauges: make(map[string]float64),
	Counters: make(map[string]int64),
}
// Мапа для хранения имеющегося в структуре counter
var countersMap = map[string]int64{}
var gaugesMap = map[string]float64{}

func main() {
	if err := logger.Initialize(flagLogLevel); err != nil{
		log.Fatal(err)
	}

	r := chi.NewRouter()

	r.Get("/", logger.WithLogging(giveHTML))
	r.Post("/update/", logger.WithLogging(handleSendJSON))
	r.Post("/value/", logger.WithLogging(handleGetJson))
	r.Post("/update/counter/{name}/{value}", logger.WithLogging(handleCounter))
	r.Post("/update/gauge/{name}/{value}", logger.WithLogging(handleGauge))
	r.Post("/update/{metricType}/{name}/{value}", logger.WithLogging(handleWrongType))
	r.Get("/value/{metricType}/{name}", logger.WithLogging(giveValue))

	parseFlagsServer()

	logger.Log.Info("Server is running on addr", zap.String("addr", flagRunAddr))
	log.Fatal(http.ListenAndServe(flagRunAddr, r))
}

func handleGetJson(w http.ResponseWriter, r *http.Request){
	var metrics models.Metrics

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&metrics); err != nil{
		logger.Log.Info("decoding request JSON body error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}


	response := models.Metrics{
		ID: metrics.ID,
		MType: metrics.MType,
	}
	switch response.MType{
	case "counter":
		if _, ok := countersMap[response.ID]; !ok{
			logger.Log.Info("No counter metric with this id")
			w.WriteHeader(http.StatusNotFound)
			return
		}
		val := countersMap[response.ID]
		response.Delta = &val

		w.Header().Set("Content-type", "application/json")

		enc := json.NewEncoder(w)
		if err := enc.Encode(response); err != nil{
			logger.Log.Info("encoding response JSON body error", zap.Error(err))
			return
		}
		w.WriteHeader(http.StatusOK)

	case "gauge":
		if _, ok := gaugesMap[response.ID]; !ok{
			logger.Log.Info("No gauge metric with this id")
			w.WriteHeader(http.StatusNotFound)
			return
		}
		value := gaugesMap[response.ID]
		response.Value = &value

		w.Header().Set("Content-type", "application/json")

		enc := json.NewEncoder(w)
		if err := enc.Encode(response); err != nil{
			logger.Log.Info("encoding response JSON body error", zap.Error(err))
			return
		}
		w.WriteHeader(http.StatusOK)
	}


}

func handleSendJSON(w http.ResponseWriter, r *http.Request) {
	var metrics models.Metrics

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&metrics); err != nil{
		logger.Log.Info("decoding request JSON body error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	switch metrics.MType{
	case "gauge":
		response := models.Metrics{
			ID: metrics.ID,
			MType: metrics.MType,
			Value: metrics.Value,
		}

		if response.ID == ""{
			logger.Log.Info("Forgot metric name")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		gaugesMap[response.ID] = *metrics.Value

		enc := json.NewEncoder(w)
		if err := enc.Encode(response); err != nil{
			logger.Log.Info("encoding response JSON body error", zap.Error(err))
			return
		}

	case "counter":
		response := models.Metrics{
			ID: metrics.ID,
			MType: metrics.MType,
			Delta: metrics.Delta,
		}

		if response.ID == ""{
			logger.Log.Info("Forgot metric name")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if value, ok := countersMap[response.ID]; ok{
			newValue := *response.Delta + value
			response.Delta = &newValue
		}
		countersMap[response.ID] = *response.Delta

		enc := json.NewEncoder(w)
		if err := enc.Encode(response); err != nil{
			logger.Log.Info("encoding response JSON body error", zap.Error(err))
			return
		}
	default:
		logger.Log.Info("Wrong metric type")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)

}

func giveHTML(w http.ResponseWriter, r *http.Request){
	for name, value := range storage.Counters{
		wr := fmt.Sprintf("%s: %d\n", name, value)
		w.Write([]byte(wr))
	}

	for name, value := range storage.Gauges{
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

	storage.UpdateCounter(name, valueInt)

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

	storage.UpdateGauge(name, valueFloat)


	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

}

func giveValue(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")
	name := chi.URLParam(r, "name")

	switch metricType{
	case "counter":
		storage.GetCounter(name, w)
		return
	case "gauge":
		storage.GetGauge(name, w)
		return
	default:
		w.WriteHeader(http.StatusNotFound)
		return
	}
}

func (s *MemStorage) UpdateGauge(name string, value float64){
	s.mx.Lock()
	defer s.mx.Unlock()
	s.Gauges[name] = value
	gaugesMap[name] = value
}

func (s *MemStorage) UpdateCounter(name string, value int64){
	s.mx.Lock()
	defer s.mx.Unlock()
	s.Counters[name] += value
	countersMap[name] = s.Counters[name]
}

func (s *MemStorage) GetGauge(name string, w http.ResponseWriter){
	s.mx.Lock()
	defer s.mx.Unlock()
	if _, ok := s.Gauges[name]; !ok{
		w.WriteHeader(http.StatusNotFound)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(strconv.FormatFloat(s.Gauges[name], 'f', -1, 64)))
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
}

func (s *MemStorage) GetCounter(name string, w http.ResponseWriter){
	s.mx.Lock()
	defer s.mx.Unlock()
	if _, ok := s.Counters[name]; !ok{
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(strconv.FormatInt(s.Counters[name], 10)))
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

}
