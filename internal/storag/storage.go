package storage

import (
	"net/http"
	"strconv"
	"sync"
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
