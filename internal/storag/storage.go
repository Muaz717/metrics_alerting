package storage

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"

	"github.com/Muaz717/metrics_alerting/internal/logger"
	"go.uber.org/zap"
)

type Metrics struct{
	ID		string		`json:"id"`
	MType 	string		`json:"type"`
	Delta	*int64		`json:"delta,omitempty"`
	Value	*float64	`json:"value,omitempty"`
}

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

func (s *MemStorage) UpdateGauge(name string, value float64){
	s.mx.Lock()
	defer s.mx.Unlock()
	s.Gauges[name] = value
}

func (s *MemStorage) UpdateCounter(name string, value int64){
	s.mx.Lock()
	defer s.mx.Unlock()
	s.Counters[name] += value
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

func (s *MemStorage) UpdateGaugeJSON(metrics Metrics, w http.ResponseWriter) {
	s.mx.Lock()
	defer s.mx.Unlock()

	response := Metrics{
		ID: metrics.ID,
		MType: metrics.MType,
		Value: metrics.Value,
	}

	if response.ID == ""{
		logger.Log.Info("Forgot metric name")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if response.Value == nil{
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	s.Gauges[response.ID] = *metrics.Value

	w.Header().Set("Content-type", "application/json")
	enc := json.NewEncoder(w)
	if err := enc.Encode(response); err != nil{
		logger.Log.Info("encoding response JSON body error", zap.Error(err))
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *MemStorage) UpdateCounterJSON(metrics Metrics, w http.ResponseWriter) {
	s.mx.Lock()
	defer s.mx.Unlock()

	response := Metrics{
		ID: metrics.ID,
		MType: metrics.MType,
		Delta: metrics.Delta,
	}

	if response.ID == ""{
		logger.Log.Info("Forgot metric name")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if response.Delta == nil{
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if value, ok := s.Counters[response.ID]; ok{
		newValue := *response.Delta + value
		response.Delta = &newValue
	}
	s.Counters[response.ID] = *response.Delta

	w.Header().Set("Content-type", "application/json")
	enc := json.NewEncoder(w)
	if err := enc.Encode(response); err != nil{
		logger.Log.Info("encoding response JSON body error", zap.Error(err))
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *MemStorage) ValueGaugeJSON(metrics Metrics, w http.ResponseWriter) {
	response := Metrics{
		ID: metrics.ID,
		MType: metrics.MType,
	}

	if _, ok := s.Gauges[response.ID]; !ok{
		logger.Log.Info("No gauge metric with this id")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	value := s.Gauges[response.ID]
	response.Value = &value

	w.Header().Set("Content-type", "application/json")

	enc := json.NewEncoder(w)
	if err := enc.Encode(response); err != nil{
		logger.Log.Info("encoding response JSON body error", zap.Error(err))
		return
	}


	w.WriteHeader(http.StatusOK)
}

func (s *MemStorage) ValueCounterJSON(metrics Metrics, w http.ResponseWriter) {
	response := Metrics{
		ID: metrics.ID,
		MType: metrics.MType,
	}

	if _, ok := s.Counters[response.ID]; !ok{
		logger.Log.Info("No counter metric with this id")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	delta := s.Counters[response.ID]
	response.Delta = &delta

	w.Header().Set("Content-type", "application/json")

	enc := json.NewEncoder(w)
	if err := enc.Encode(response); err != nil{
		logger.Log.Info("encoding response JSON body error", zap.Error(err))
		return
	}

	w.WriteHeader(http.StatusOK)
}
