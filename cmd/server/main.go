package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
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
	r := chi.NewRouter()

	r.Get("/", giveHTML)
	r.Post("/update/counter/{name}/{value}", handleCounter)
	r.Post("/update/gauge/{name}/{value}", handleGauge)
	r.Post("/update/{metricType}/{name}/{value}", handleWrongType)
	r.Get("/value/{metricType}/{name}", giveValue)

	parseFlagsServer()

	log.Fatal(http.ListenAndServe(flagRunAddr, r))
}

func giveHTML(w http.ResponseWriter, r *http.Request){
	for name, value := range storage.Counters{
		fmt.Fprintf(w, "%s: %d\n", name, value)
	}

	for name, value := range storage.Gauges{
		fmt.Fprintf(w, "%s: %f\n", name, value)
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
		if _, ok := storage.Counters[name]; !ok{
			w.WriteHeader(http.StatusNotFound)
			return
		}

		io.WriteString(w, strconv.FormatInt(storage.Counters[name], 10))
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		return

	case "gauge":
		if _, ok := storage.Gauges[name]; !ok{
			w.WriteHeader(http.StatusNotFound)
			return
		}

		io.WriteString(w, strconv.FormatFloat(storage.Gauges[name], 'f', -1, 64))
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		return
	default:
		w.WriteHeader(http.StatusNotFound)
		return
	}
}

func (s *MemStorage) UpdateGauge(name string, value float64){
	s.Gauges[name] = value
}

func (s *MemStorage) UpdateCounter(name string, value int64){
	s.Counters[name] += value
}
