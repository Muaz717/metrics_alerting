package main

import (
	"net/http"
	"strconv"
	"strings"
)

// Тип хранилища для метрик
type MemStorage struct {
	Gauges map[string]float64
	Counters map[string]int64
}

// Интерфейс для взаимодействия с хранилищем
type Storage interface {
	UpdateGauge(string, float64)
	UpdateCounter(string, int64)
}

// Инициализация хранилища
var storage = &MemStorage{
	Gauges: make(map[string]float64),
	Counters: make(map[string]int64),
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/update/", handleMetric)
	mux.HandleFunc("/", handleAnother)

	err := http.ListenAndServe(":8080", mux)
	if err != nil{
		panic(err)
	}
}

// Обработчик POST-запросов для приёма метрик
func handleMetric(w http.ResponseWriter, r *http.Request) {
	path := strings.Split(r.URL.Path, "/")

	if len(path) != 4 && len(path) != 5 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	metricType := path[2]

	// Проверка типа метрики
	switch metricType{
	case "gauge":
		if len(path) == 4 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		value, err := strconv.ParseFloat(path[4], 64)
		if err != nil{
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Обновление метрики типа "gauge"
		storage.UpdateGauge(path[3], value)

	case "counter":
		if len(path) == 4 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		value, err := strconv.ParseInt(path[4], 10, 64)
		if err != nil{
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Обновление метрики типа "counter"
		storage.UpdateCounter(path[3], value)

	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Отправка ответа 200 OK
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

}

// Обработчик всех остальных запросов
func handleAnother(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w,r)
}

func (s *MemStorage) UpdateGauge(name string, value float64){
	s.Gauges[name] = value
}

func (s *MemStorage) UpdateCounter(name string, value int64){
	s.Counters[name] += value
}
