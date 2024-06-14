package main

import (
	"encoding/json"
	"fmt"
	"time"

	// "fmt"
	"log"
	"net/http"
	"sync"

	"github.com/Muaz717/metrics_alerting/internal/config"
	gziper "github.com/Muaz717/metrics_alerting/internal/gzip"
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
var store storage.Storage

func main() {
	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Use(logger.WithLogging)

		r.Get("/", gziper.GzipMiddleware(getHTML))
		r.Post("/update/", gziper.GzipMiddleware(handleUpdateJSON))
		r.Post("/value/", gziper.GzipMiddleware(handleValueJSON))

		r.Post("/update/{metricType}/{name}/{value}", handleMetric)
		r.Get("/value/{metricType}/{name}", getValue)
	})

	if err := logger.Initialize("info"); err != nil{
		log.Fatal(err)
	}

	cfg, err := config.NewServerConfig()
	if err != nil{
		log.Fatal(err)
	}

	store, err = storage.NewStore(&cfg)
	if err != nil{
		fmt.Printf("failed to create storage: %v", err)
		// return
	}

	if cfg.StorageCfg.StoreInterval != 0 && cfg.StorageCfg.FileStoragePath != "" {
		go func() {
			tickerStore := time.NewTicker(time.Duration(cfg.StorageCfg.StoreInterval) * time.Second)
			for range tickerStore.C {
				err := storage.SaveMetrics(store, cfg.StorageCfg.FileStoragePath)
				if err != nil {
					log.Fatal("failed to save metrics: %w", err)
				}
			}
		}()
	}

	logger.Log.Info("Server is running on addr", zap.String("addr", cfg.Host))
	log.Fatal(http.ListenAndServe(cfg.Host, r))
}

func handleValueJSON(w http.ResponseWriter, r *http.Request){
	var metrics storage.Metrics

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&metrics); err != nil{
		logger.Log.Info("decoding request JSON body error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response, err := store.GetJSON(metrics)
	if err != nil{
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)

	enc := json.NewEncoder(w)
	if err := enc.Encode(response); err != nil{
		logger.Log.Info("encoding response JSON body error", zap.Error(err))
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

	response, err := store.SaveJSON(metrics)
	if err != nil{
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)

	enc := json.NewEncoder(w)
	if err := enc.Encode(response); err != nil{
		logger.Log.Info("encoding response JSON body error", zap.Error(err))
		return
	}
}

func handleMetric(w http.ResponseWriter, r *http.Request){
	mType := chi.URLParam(r, "metricType")
	mName := chi.URLParam(r, "name")
	mValue := chi.URLParam(r, "value")

	err := metricsStorage.SaveMetric(mType, mName, mValue)
	if err != nil{
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

func getHTML(w http.ResponseWriter, r *http.Request){
	mx.Lock()
	defer mx.Unlock()
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	html := store.GetAllMetrics()
	w.Write([]byte(html))
}

func getValue(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")
	name := chi.URLParam(r, "name")

	mValue, err := store.GetMetric(metricType, name)
	if err != nil{
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	w.Write([]byte(mValue))
}
